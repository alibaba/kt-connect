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
)

func AutoMesh(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// 1. Get service to mesh
	svcName, err := general.GetServiceByResourceName(ctx, k, resourceName, opts)
	if err != nil {
		return err
	}

	// 2. Lock service to avoid conflict
	//if err = general.LockService(ctx, k, svcName, opts.Namespace, 0); err != nil {
	//	return err
	//}
	//defer general.UnlockService(ctx, k, svcName, opts.Namespace)

	// 3. Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opts.MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opts.RuntimeOptions.Mesh = versionMark

	svc, err := k.GetService(ctx, svcName, opts.Namespace)
	if err != nil {
		return err
	}
	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	// 4. Create origin service
	originSvcName := svcName + common.OriginServiceSuffix
	if err = createOriginService(ctx, k, originSvcName, ports, svc.Spec.Selector, opts); err != nil {
		return err
	}

	// 5. Create shadow service
	shadowName := svcName + common.MeshPodInfix + meshVersion
	shadowLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtName: shadowName,
	}
	if err = createShadowService(ctx, k, shadowName, ports, shadowLabels, opts); err != nil {
		return err
	}

	// 6. Create router pod
	routerPodName := svcName + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtName: routerPodName,
	}
	if err = createRouter(ctx, k, routerPodName, svcName, ports, routerLabels, versionMark, opts); err != nil {
		return err
	}

	// 7. Let target service select router pod
	if err = general.UpdateServiceSelector(ctx, k, svc, routerLabels); err != nil {
		return err
	}

	// 8. Create shadow pod
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
