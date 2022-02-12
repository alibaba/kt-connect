package mesh

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strconv"
	"strings"
	"time"
)

func AutoMesh(resourceName string) error {
	// Get service to mesh
	svc, err := general.GetServiceByResourceName(resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}
	if port := util.FindInvalidRemotePort(opt.Get().MeshOptions.Expose, svc.Spec.Ports); port != "" {
		return fmt.Errorf("target port %s not exists in service %s", port, svc.Name)
	}

	// Lock service to avoid conflict, must be first step
	svc, err = general.LockService(svc.Name, opt.Get().Namespace, 0)
	if err != nil {
		return err
	}
	defer general.UnlockService(svc.Name, opt.Get().Namespace)

	if svc.Annotations != nil && svc.Annotations[common.KtSelector] != "" && svc.Spec.Selector[common.KtRole] == common.RoleExchangeShadow {
		return fmt.Errorf("another user is exchanging service '%s', cannot apply mesh", svc.Name)
	}

	// Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opt.Get().MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opt.Get().RuntimeStore.Mesh = versionMark

	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	// Check name usable
	if err = isNameUsable(svc.Name, meshVersion, 0); err != nil {
		return err
	}

	// Create origin service
	if err = createStuntmanService(svc, ports); err != nil {
		return err
	}

	// Create shadow service
	shadowName := svc.Name + common.MeshPodInfix + meshVersion
	targetMark := util.RandomString(20)
	shadowLabels := map[string]string{
		common.KtRole:   common.RoleMeshShadow,
		common.KtTarget: targetMark,
	}
	if err = createShadowService(shadowName, ports, shadowLabels); err != nil {
		return err
	}

	// Create router pod
	// Must after origin service and shadow service, otherwise will cause 'host not found in upstream' error
	routerPodName := svc.Name + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtRole:   common.RoleRouter,
		common.KtTarget: targetMark,
	}
	if err = createRouter(routerPodName, svc.Name, ports, routerLabels, versionMark); err != nil {
		return err
	}

	// Let target service select router pod
	// Must after router pod created, otherwise request will be interrupted
	if err = general.UpdateServiceSelector(svc.Name, opt.Get().Namespace, routerLabels); err != nil {
		return err
	}

	// Create shadow pod
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", shadowName),
	}
	if err = general.CreateShadowAndInbound(shadowName, opt.Get().MeshOptions.Expose, shadowLabels, annotations); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header '%s: %s' ", strings.ToUpper(meshKey), meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func isNameUsable(name, meshVersion string, times int) error {
	if times > 10 {
		return fmt.Errorf("meshing pod for service %s still terminating, please try again later", name)
	}
	shadowName := name + common.MeshPodInfix + meshVersion
	if pod, err := cluster.Ins().GetPod(shadowName, opt.Get().Namespace); err == nil {
		if pod.DeletionTimestamp == nil {
			msg := fmt.Sprintf("Another user is meshing service '%s' via version '%s'", name, meshVersion)
			if opt.Get().MeshOptions.VersionMark != "" {
				return fmt.Errorf("%s, please specify a different version mark", msg)
			}
			return fmt.Errorf( "%s, please retry or use '--versionMark' parameter to spcify an uniq one", msg)
		}
		log.Info().Msgf("Previous meshing pod for service '%s' not finished yet, waiting ...", name)
		time.Sleep(3 * time.Second)
		return isNameUsable(name, meshVersion, times + 1)
	}
	return nil
}

func createShadowService(shadowSvcName string, ports map[int]int,
	selectors map[string]string) error {
	if _, err := cluster.Ins().CreateService(&cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        shadowSvcName,
			Namespace:   opt.Get().Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		External:  false,
		Ports:     ports,
		Selectors: selectors,
	}); err != nil {
		return err
	}

	opt.Get().RuntimeStore.Service = shadowSvcName
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func createRouter(routerPodName string, svcName string, ports map[int]int, labels map[string]string, versionMark string) error {
	namespace := opt.Get().Namespace
	routerLabels := util.MergeMap(labels, map[string]string{common.ControlBy: common.KubernetesTool})
	routerPod, err := cluster.Ins().GetPod(routerPodName, namespace)
	if err == nil && routerPod.DeletionTimestamp != nil {
		routerPod, err = cluster.Ins().WaitPodTerminate(routerPodName, namespace)
	}
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			// Failed to get or wait router pod
			return err
		}
		// Router not exist or just terminated
		annotations := map[string]string{common.KtRefCount: "1", common.KtConfig: fmt.Sprintf("service=%s", svcName)}
		if err = cluster.CreateRouterPod(routerPodName, routerLabels, annotations); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(common.DefaultContainer, routerPodName, namespace,
			common.RouterBin, "setup", svcName, toPortMapParameter(ports), versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	} else {
		// Router pod exist
		cluster.Ins().UpdatePodHeartBeat(routerPodName, namespace)
		if _, err = strconv.Atoi(routerPod.Annotations[common.KtRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = cluster.Ins().IncreaseRef(routerPodName, namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(common.DefaultContainer, routerPodName, namespace,
			common.RouterBin, "add", versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	}
	log.Info().Msgf("Router pod configuration done")
	opt.Get().RuntimeStore.Router = routerPodName
	return nil
}

func createStuntmanService(svc *coreV1.Service, ports map[int]int) error {
	stuntmanSvcName := svc.Name + common.StuntmanServiceSuffix
	namespace := opt.Get().Namespace
	if stuntmanSvc, err := cluster.Ins().GetService(stuntmanSvcName, namespace); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		if svc.Annotations != nil && svc.Annotations[common.KtSelector] != "" {
			return fmt.Errorf("service %s should not have %s annotation, please try use 'ktctl clean' to restore it",
				svc.Name, common.KtSelector)
		}
		if _, err = cluster.Ins().CreateService(&cluster.SvcMetaAndSpec{
			Meta: &cluster.ResourceMeta{
				Name:        stuntmanSvcName,
				Namespace:   namespace,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			External:  false,
			Ports:     ports,
			Selectors: svc.Spec.Selector,
		}); err != nil {
			return err
		}
		log.Info().Msgf("Service %s created", stuntmanSvcName)
	} else if stuntmanSvc.Annotations == nil || stuntmanSvc.Annotations[common.ControlBy] != common.KubernetesTool {
		return fmt.Errorf("service %s exists, but not created by kt", stuntmanSvcName)
	} else {
		cluster.Ins().UpdateServiceHeartBeat(stuntmanSvcName, namespace)
		log.Info().Msgf("Origin service already exists")
	}
	return nil
}

func toPortMapParameter(ports map[int]int) string {
	// input: { 80:8080, 70:7000 }
	// output: "80:8080,70:7000"
	if len(ports) == 0 {
		return ""
	}
	s := ""
	for k, v := range ports {
		s = s + "," + strconv.Itoa(k) + ":" + strconv.Itoa(v)
	}
	return s[1:]
}
