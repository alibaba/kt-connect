package mesh

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strconv"
	"strings"
	"time"
)

func AutoMesh(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// Get service to mesh
	svc, err := general.GetServiceByResourceName(ctx, k, resourceName, opts.Namespace)
	if err != nil {
		return err
	}

	// Lock service to avoid conflict
	err = general.LockService(ctx, k, svc.Name, opts.Namespace, 0)
	if err != nil {
		return err
	}
	defer general.UnlockService(ctx, k, svc.Name, opts.Namespace)

	// Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opts.MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opts.RuntimeOptions.Mesh = versionMark

	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	// Check name usable
	if err = isNameUsable(ctx, k, svc.Name, meshVersion, opts, 0); err != nil {
		return err
	}

	// Create origin service
	originSvcName := svc.Name + common.OriginServiceSuffix
	if err = createOriginService(ctx, k, originSvcName, ports, svc.Spec.Selector, opts); err != nil {
		return err
	}

	// Create shadow service
	shadowName := svc.Name + common.MeshPodInfix + meshVersion
	shadowLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtName: shadowName,
	}
	if err = createShadowService(ctx, k, shadowName, ports, shadowLabels, opts); err != nil {
		return err
	}

	// Create router pod
	// Must after origin service and shadow service, otherwise will cause 'host not found in upstream' error
	routerPodName := svc.Name + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtRole: common.RoleRouter,
		common.KtName: routerPodName,
	}
	if err = createRouter(ctx, k, routerPodName, svc.Name, ports, routerLabels, versionMark, opts); err != nil {
		return err
	}

	// Let target service select router pod
	// Must after router pod created, otherwise request will be interrupted
	if err = general.UpdateServiceSelector(ctx, k, svc.Name, opts.Namespace, routerLabels); err != nil {
		return err
	}

	// Create shadow pod
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", shadowName),
	}
	if err = general.CreateShadowAndInbound(ctx, k, shadowName, opts.MeshOptions.Expose, shadowLabels, annotations, opts); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header '%s: %s' ", strings.ToUpper(meshKey), meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func isNameUsable(ctx context.Context, k cluster.KubernetesInterface, name, meshVersion string,
	opts *options.DaemonOptions, times int) error {
	if times > 10 {
		return fmt.Errorf("meshing pod for service %s still terminating, please try again later", name)
	}
	shadowName := name + common.MeshPodInfix + meshVersion
	if pod, err := k.GetPod(ctx, shadowName, opts.Namespace); err == nil {
		if pod.DeletionTimestamp == nil {
			msg := fmt.Sprintf("Another user is meshing service '%s' via version '%s'", name, meshVersion)
			if opts.MeshOptions.VersionMark != "" {
				return fmt.Errorf("%s, please specify a different version mark", msg)
			}
			return fmt.Errorf( "%s, please retry or use '--versionMark' parameter to spcify an uniq one", msg)
		}
		log.Info().Msgf("Previous meshing pod for service '%s' not finished yet, waiting ...", name)
		time.Sleep(3 * time.Second)
		return isNameUsable(ctx, k, name, meshVersion, opts, times + 1)
	}
	return nil
}

func createShadowService(ctx context.Context, k cluster.KubernetesInterface, shadowSvcName string, ports map[int]int,
	selectors map[string]string, options *options.DaemonOptions) error {
	if _, err := k.CreateService(ctx, &cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        shadowSvcName,
			Namespace:   options.Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		External:  false,
		Ports:     ports,
		Selectors: selectors,
	}); err != nil {
		return err
	}

	options.RuntimeOptions.Service = shadowSvcName
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func createRouter(ctx context.Context, k cluster.KubernetesInterface, routerPodName string, svcName string,
	ports map[int]int, labels map[string]string, versionMark string, opts *options.DaemonOptions) error {
	routerLabels := util.MergeMap(labels, map[string]string{common.ControlBy: common.KubernetesTool})
	routerPod, err := k.GetPod(ctx, routerPodName, opts.Namespace)
	if err == nil && routerPod.DeletionTimestamp != nil {
		routerPod, err = k.WaitPodTerminate(ctx, routerPodName, opts.Namespace)
	}
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		annotations := map[string]string{common.KtRefCount: "1", common.KtConfig: fmt.Sprintf("service=%s", svcName)}
		if err = cluster.CreateRouterPod(ctx, k, routerPodName, opts, routerLabels, annotations); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		stdout, stderr, err2 := k.ExecInPod(common.DefaultContainer, routerPodName, opts.Namespace, *opts.RuntimeOptions,
			common.RouterBin, "setup", svcName, toPortMapParameter(ports), versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	} else {
		if _, err = strconv.Atoi(routerPod.Annotations[common.KtRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = k.IncreaseRef(ctx, routerPodName, opts.Namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		stdout, stderr, err2 := k.ExecInPod(common.DefaultContainer, routerPodName, opts.Namespace, *opts.RuntimeOptions,
			common.RouterBin, "add", versionMark)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	}
	log.Info().Msgf("Router pod configuration done")
	opts.RuntimeOptions.Router = routerPodName
	return nil
}

func createOriginService(ctx context.Context, k cluster.KubernetesInterface, originSvcName string,
	ports map[int]int, selectors map[string]string, options *options.DaemonOptions) error {

	_, err := k.GetService(ctx, originSvcName, options.Namespace)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		if _, err = k.CreateService(ctx, &cluster.SvcMetaAndSpec{
			Meta: &cluster.ResourceMeta{
				Name:        originSvcName,
				Namespace:   options.Namespace,
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
