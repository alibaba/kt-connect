package mesh

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
	"strings"
	"time"
)

func AutoMesh(svc *coreV1.Service) error {
	// Lock service to avoid conflict, must be first step
	svc, err := general.LockService(svc.Name, opt.Get().Global.Namespace, 0)
	if err != nil {
		return err
	}
	defer general.UnlockService(svc.Name, opt.Get().Global.Namespace)

	if svc.Annotations != nil && svc.Annotations[util.KtSelector] != "" && svc.Spec.Selector[util.KtRole] == util.RoleExchangeShadow {
		return fmt.Errorf("another user%s is exchanging service '%s', cannot apply mesh",
			general.GetOccupiedUser(svc.Spec.Selector), svc.Name)
	}

	// Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opt.Get().Mesh.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opt.Store.Mesh = versionMark

	portToNames := general.GetTargetPorts(svc)
	ports := make(map[int]int)
	for _, specPort := range svc.Spec.Ports {
		if specPort.TargetPort.Type == intstr.Int {
			ports[int(specPort.Port)] = specPort.TargetPort.IntValue()
		} else {
			podPort := -1
			for p, n := range portToNames {
				if n == specPort.TargetPort.StrVal {
					podPort = p
					break
				}
			}
			if podPort < 0 {
				return fmt.Errorf("cannot found port number of target port '%s' of service %s",
					specPort.TargetPort.StrVal, svc.Name)
			}
			ports[int(specPort.Port)] = podPort
		}
	}

	// Check name usable
	if err = isNameUsable(svc.Name, meshVersion, 0); err != nil {
		return err
	}

	// Check service in sanity status
	if err = sanityCheck(svc); err != nil {
		return err
	}

	// Create stuntman service
	if err = createStuntmanService(svc, ports); err != nil {
		return err
	}

	// Create shadow service
	shadowName := svc.Name + util.MeshPodInfix + meshVersion
	shadowLabels := map[string]string{
		util.KtRole:   util.RoleMeshShadow,
		util.KtTarget: util.RandomString(20),
	}
	if err = createShadowService(shadowName, ports, shadowLabels); err != nil {
		return err
	}

	// Create router pod
	// Must after stuntman service and shadow service, otherwise will cause 'host not found in upstream' error
	routerPodName := svc.Name + util.RouterPodSuffix
	routerLabels := map[string]string{
		util.KtRole:   util.RoleRouter,
	}
	if err = createRouter(routerPodName, svc.Name, ports, routerLabels, versionMark); err != nil {
		return err
	}

	// Let target service select router pod
	// Must after router pod created, otherwise request will be interrupted
	if err = general.UpdateServiceSelector(svc.Name, opt.Get().Global.Namespace, routerLabels); err != nil {
		return err
	}
	opt.Store.Origin = svc.Name

	// Create shadow pod
	annotations := map[string]string{
		util.KtConfig: fmt.Sprintf("service=%s", shadowName),
	}
	if err = general.CreateShadowAndInbound(shadowName, opt.Get().Mesh.Expose,
		shadowLabels, annotations, portToNames); err != nil {
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
	shadowName := name + util.MeshPodInfix + meshVersion
	if pod, err := cluster.Ins().GetPod(shadowName, opt.Get().Global.Namespace); err == nil {
		if pod.DeletionTimestamp == nil {
			msg := fmt.Sprintf("Another user is meshing service '%s' via version '%s'", name, meshVersion)
			if opt.Get().Mesh.VersionMark != "" {
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

func sanityCheck(svc *coreV1.Service) error {
	if svc.Annotations != nil && svc.Annotations[util.KtSelector] != "" {
		return fmt.Errorf("service %s should not have %s annotation, please try use 'ktctl recover %s' to restore it",
			svc.Name, util.KtSelector, svc.Name)
	} else if svc.Spec.Selector[util.KtRole] != "" {
		return fmt.Errorf("service %s should not point to kt pods, please try use 'ktctl recover %s' to restore it",
			svc.Name, svc.Name)
	}
	return nil
}

func createShadowService(shadowSvcName string, ports map[int]int,
	selectors map[string]string) error {
	if _, err := cluster.Ins().CreateService(&cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        shadowSvcName,
			Namespace:   opt.Get().Global.Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		External:  false,
		Ports:     ports,
		Selectors: selectors,
	}); err != nil {
		return err
	}

	opt.Store.Service = shadowSvcName
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func createRouter(routerPodName string, svcName string, ports map[int]int, labels map[string]string, versionMark string) error {
	namespace := opt.Get().Global.Namespace
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
		labels[util.KtTarget] = util.RandomString(20)
		annotations := map[string]string{util.KtRefCount: "1", util.KtConfig: fmt.Sprintf("service=%s", svcName)}
		if _, err = cluster.Ins().CreateRouterPod(routerPodName, labels, annotations, ports); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(util.DefaultContainer, routerPodName, namespace,
			util.RouterBin, "setup", svcName, toPortMapParameter(ports), versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	} else {
		// Router pod exist
		labels[util.KtTarget] = routerPod.Labels[util.KtTarget]
		cluster.Ins().UpdatePodHeartBeat(routerPodName, namespace)
		if _, err = strconv.Atoi(routerPod.Annotations[util.KtRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = cluster.Ins().IncreasePodRef(routerPodName, namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(util.DefaultContainer, routerPodName, namespace,
			util.RouterBin, "add", versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	}
	log.Info().Msgf("Router pod configuration done")
	opt.Store.Router = routerPodName
	return nil
}

func createStuntmanService(svc *coreV1.Service, ports map[int]int) error {
	stuntmanSvcName := svc.Name + util.StuntmanServiceSuffix
	namespace := opt.Get().Global.Namespace
	if stuntmanSvc, err := cluster.Ins().GetService(stuntmanSvcName, namespace); err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		} else if _, err = cluster.Ins().CreateService(&cluster.SvcMetaAndSpec{
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
	} else if stuntmanSvc.Labels[util.ControlBy] != util.KubernetesToolkit {
		return fmt.Errorf("service %s exists, but not created by kt", stuntmanSvcName)
	} else {
		cluster.Ins().UpdateServiceHeartBeat(stuntmanSvcName, namespace)
		log.Info().Msgf("Stuntman service already exists")
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
