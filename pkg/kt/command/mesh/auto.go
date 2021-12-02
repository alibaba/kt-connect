package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"strconv"
	"strings"
	"time"
)

func AutoMesh(ctx context.Context, k cluster.KubernetesInterface, deploymentName string, opts *options.DaemonOptions) error {
	app, err := lockAndFetchDeployment(ctx, k, deploymentName, opts.Namespace, 0)
	if err != nil {
		return err
	}
	defer unlockDeployment(ctx, k, deploymentName, opts.Namespace)

	svc, err := getServiceByDeployment(ctx, k, app, opts)
	if err != nil {
		return err
	}

	meshKey, meshVersion := getVersion(opts.MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opts.RuntimeOptions.Mesh = versionMark

	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	originSvcName := svc.Name + common.OriginServiceSuffix
	if err = createOriginService(ctx, k, originSvcName, ports, app.Spec.Selector.MatchLabels, opts); err != nil {
		return err
	}

	shadowPodName := deploymentName + common.MeshPodInfix + meshVersion
	shadowSvcName := svc.Name + common.MeshPodInfix + meshVersion
	shadowLabels := map[string]string{
		common.KtComponent: common.ComponentMesh,
		common.KtRole: common.RoleShadow,
		common.KtName: shadowPodName,
	}
	if err = createShadowService(ctx, k, shadowSvcName, ports, util.CopyMap(shadowLabels), opts); err != nil {
		return err
	}

	routerPodName := deploymentName + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtComponent: common.ComponentMesh,
		common.KtRole: common.RoleRouter,
		common.KtName: routerPodName,
	}
	if err = createRouter(ctx, k, routerPodName, svc.Name, ports, util.CopyMap(routerLabels), versionMark, opts); err != nil {
		return err
	}

	if _, ok := svc.Annotations[common.KtSelector]; !ok {
		if marshaledSelector, err2 := json.Marshal(svc.Spec.Selector); err2 != nil {
			log.Error().Err(err).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err2
		} else {
			util.MapPut(svc.Annotations, common.KtSelector, string(marshaledSelector))
		}
	}
	svc.Spec.Selector = routerLabels
	if _, err = k.UpdateService(ctx, svc); err != nil {
		return err
	}

	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", shadowSvcName),
	}
	if err = createShadowAndInbound(ctx, k, shadowPodName, shadowLabels, annotations, opts); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header '%s: %s' ", strings.ToUpper(meshKey), meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func lockAndFetchDeployment(ctx context.Context, k cluster.KubernetesInterface, deploymentName, namespace string, times int) (*appV1.Deployment, error) {
	if times > 10 {
		log.Warn().Msgf("Unable to obtain auto mesh lock, please try again later.")
		return nil, fmt.Errorf("failed to obtain auto meth lock of deployment %s", deploymentName)
	}
	app, err := k.GetDeployment(ctx, deploymentName, namespace)
	if err != nil {
		return nil, err
	}

	if _, ok := app.Annotations[common.KtLock]; ok {
		log.Info().Msgf("Another user is meshing deployment %s, waiting for lock ...", deploymentName)
		time.Sleep(3 * time.Second)
		return lockAndFetchDeployment(ctx, k, deploymentName, namespace, times + 1)
	} else {
		util.MapPut(app.Annotations, common.KtLock, util.GetTimestamp())
		if app, err = k.UpdateDeployment(ctx, app); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock deployment %s", deploymentName)
			return lockAndFetchDeployment(ctx, k, deploymentName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Deployment %s locked for auto mesh", deploymentName)
	return app, nil
}

func unlockDeployment(ctx context.Context, k cluster.KubernetesInterface, deploymentName, namespace string) {
	app, err := k.GetDeployment(ctx, deploymentName, namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get deployment %s for unlock", app.Name)
		return
	}
	delete(app.Annotations, common.KtLock)
	if _, err := k.UpdateDeployment(ctx, app); err != nil {
		log.Warn().Err(err).Msgf("Failed to unlock deployment %s", app.Name)
	}
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

func getServiceByDeployment(ctx context.Context, k cluster.KubernetesInterface, app *appV1.Deployment,
	options *options.DaemonOptions) (*coreV1.Service, error) {
	svcList, err := k.GetServices(ctx, app.Spec.Selector.MatchLabels, options.Namespace)
	if err != nil {
		return nil, err
	} else if len(svcList) == 0 {
		return nil, fmt.Errorf("failed to find service for deployment \"%s\", with labels \"%v\"",
			app.Name, app.Spec.Selector.MatchLabels)
	} else if len(svcList) > 1 {
		svcNames := svcList[0].Name
		for i, svc := range svcList {
			if i > 0 {
				svcNames = svcNames + ", " + svc.Name
			}
		}
		log.Warn().Msgf("Found %d services match deployment \"%s\": %s. First one will be used.",
			len(svcList), app.Name, svcNames)
	}
	svc := svcList[0]
	if strings.HasSuffix(svc.Name, common.OriginServiceSuffix) {
		return k.GetService(ctx, strings.TrimSuffix(svc.Name, common.OriginServiceSuffix), options.Namespace)
	}
	return &svc, nil
}

func createRouter(ctx context.Context, k cluster.KubernetesInterface, routerPodName string, svcName string,
	ports map[int]int, routerLabels map[string]string, versionMark string, opts *options.DaemonOptions) error {
	routerPod, err := k.GetPod(ctx, routerPodName, opts.Namespace)
	routerLabels[common.ControlBy] = common.KubernetesTool
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
