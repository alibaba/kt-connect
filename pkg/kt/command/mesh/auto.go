package mesh

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strconv"
	"strings"
	"time"
)

func AutoMesh(ctx context.Context, resourceName string) error {
	// Get service to mesh
	svc, err := general.GetServiceByResourceName(ctx, resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}

	// Lock service to avoid conflict, must be first step
	svc, err = general.LockService(ctx, svc.Name, opt.Get().Namespace, 0)
	if err != nil {
		return err
	}
	defer general.UnlockService(ctx, svc.Name, opt.Get().Namespace)

	if svc.Annotations != nil && svc.Annotations[common.KtSelector] != "" && svc.Spec.Selector[common.KtRole] == common.RoleExchangeShadow {
		return fmt.Errorf("another user is exchanging service '%s', cannot apply mesh", svc.Name)
	}

	// Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opt.Get().MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opt.Get().RuntimeOptions.Mesh = versionMark

	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	// Check name usable
	if err = isNameUsable(ctx, svc.Name, meshVersion, 0); err != nil {
		return err
	}

	// Create origin service
	originSvcName := svc.Name + common.OriginServiceSuffix
	if err = createOriginService(ctx, originSvcName, ports, svc.Spec.Selector); err != nil {
		return err
	}

	// Create shadow service
	shadowName := svc.Name + common.MeshPodInfix + meshVersion
	targetMark := util.RandomString(20)
	shadowLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtTarget: targetMark,
	}
	if err = createShadowService(ctx, shadowName, ports, shadowLabels); err != nil {
		return err
	}

	// Create router pod
	// Must after origin service and shadow service, otherwise will cause 'host not found in upstream' error
	routerPodName := svc.Name + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtRole: common.RoleRouter,
		common.KtTarget: targetMark,
	}
	if err = createRouter(ctx, routerPodName, svc.Name, ports, routerLabels, versionMark); err != nil {
		return err
	}

	// Let target service select router pod
	// Must after router pod created, otherwise request will be interrupted
	if err = general.UpdateServiceSelector(ctx, svc.Name, opt.Get().Namespace, routerLabels); err != nil {
		return err
	}

	// Create shadow pod
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", shadowName),
	}
	if err = general.CreateShadowAndInbound(ctx, shadowName, opt.Get().MeshOptions.Expose, shadowLabels, annotations); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header '%s: %s' ", strings.ToUpper(meshKey), meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func isNameUsable(ctx context.Context, name, meshVersion string, times int) error {
	if times > 10 {
		return fmt.Errorf("meshing pod for service %s still terminating, please try again later", name)
	}
	shadowName := name + common.MeshPodInfix + meshVersion
	if pod, err := cluster.Ins().GetPod(ctx, shadowName, opt.Get().Namespace); err == nil {
		if pod.DeletionTimestamp == nil {
			msg := fmt.Sprintf("Another user is meshing service '%s' via version '%s'", name, meshVersion)
			if opt.Get().MeshOptions.VersionMark != "" {
				return fmt.Errorf("%s, please specify a different version mark", msg)
			}
			return fmt.Errorf( "%s, please retry or use '--versionMark' parameter to spcify an uniq one", msg)
		}
		log.Info().Msgf("Previous meshing pod for service '%s' not finished yet, waiting ...", name)
		time.Sleep(3 * time.Second)
		return isNameUsable(ctx, name, meshVersion, times + 1)
	}
	return nil
}

func createShadowService(ctx context.Context, shadowSvcName string, ports map[int]int,
	selectors map[string]string) error {
	if _, err := cluster.Ins().CreateService(ctx, &cluster.SvcMetaAndSpec{
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

	opt.Get().RuntimeOptions.Service = shadowSvcName
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func createRouter(ctx context.Context, routerPodName string, svcName string,
	ports map[int]int, labels map[string]string, versionMark string) error {
	routerLabels := util.MergeMap(labels, map[string]string{common.ControlBy: common.KubernetesTool})
	routerPod, err := cluster.Ins().GetPod(ctx, routerPodName, opt.Get().Namespace)
	if err == nil && routerPod.DeletionTimestamp != nil {
		routerPod, err = cluster.Ins().WaitPodTerminate(ctx, routerPodName, opt.Get().Namespace)
	}
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			// Failed to get or wait router pod
			return err
		}
		// Router not exist or just terminated
		annotations := map[string]string{common.KtRefCount: "1", common.KtConfig: fmt.Sprintf("service=%s", svcName)}
		if err = cluster.CreateRouterPod(ctx, routerPodName, routerLabels, annotations); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(common.DefaultContainer, routerPodName, opt.Get().Namespace,
			common.RouterBin, "setup", svcName, toPortMapParameter(ports), versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	} else {
		// Router pod exist
		if _, err = strconv.Atoi(routerPod.Annotations[common.KtRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = cluster.Ins().IncreaseRef(ctx, routerPodName, opt.Get().Namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		stdout, stderr, err2 := cluster.Ins().ExecInPod(common.DefaultContainer, routerPodName, opt.Get().Namespace,
			common.RouterBin, "add", versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	}
	log.Info().Msgf("Router pod configuration done")
	opt.Get().RuntimeOptions.Router = routerPodName
	return nil
}

func createOriginService(ctx context.Context, originSvcName string,
	ports map[int]int, selectors map[string]string) error {

	_, err := cluster.Ins().GetService(ctx, originSvcName, opt.Get().Namespace)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		if _, err = cluster.Ins().CreateService(ctx, &cluster.SvcMetaAndSpec{
			Meta: &cluster.ResourceMeta{
				Name:        originSvcName,
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
		log.Info().Msgf("Service %s created", originSvcName)
	} else {
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
