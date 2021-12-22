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

func AutoMesh(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// 1.Get service to mesh
	svcName, err := getServiceToMesh(ctx, k, resourceName, opts)
	if err != nil {
		return err
	}

	// 2.Lock service to avoid conflict
	svc, err := lockAndFetchService(ctx, k, svcName, opts.Namespace, 0)
	if err != nil {
		return err
	}
	defer unlockService(ctx, k, svcName, opts.Namespace)

	// 3.Parse or generate mesh kv
	meshKey, meshVersion := getVersion(opts.MeshOptions.VersionMark)
	versionMark := meshKey + ":" + meshVersion
	opts.RuntimeOptions.Mesh = versionMark

	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	// 4.Create origin service
	originSvcName := svc.Name + common.OriginServiceSuffix
	if err = createOriginService(ctx, k, originSvcName, ports, svc.Spec.Selector, opts); err != nil {
		return err
	}

	// 5.Create shadow service
	shadowName := svc.Name + common.MeshPodInfix + meshVersion
	shadowLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtName: shadowName,
	}
	if err = createShadowService(ctx, k, shadowName, ports, util.CopyMap(shadowLabels), opts); err != nil {
		return err
	}

	// 6.Create router pod
	routerPodName := svc.Name + common.RouterPodSuffix
	routerLabels := map[string]string{
		common.KtRole: common.RoleMeshShadow,
		common.KtName: routerPodName,
	}
	if err = createRouter(ctx, k, routerPodName, svc.Name, ports, util.CopyMap(routerLabels), versionMark, opts); err != nil {
		return err
	}

	// 7.Let target service select router pod
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

	// 8.Create shadow pod
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", shadowName),
	}
	if err = createShadowAndInbound(ctx, k, shadowName, shadowLabels, annotations, opts); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header '%s: %s' ", strings.ToUpper(meshKey), meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func getServiceToMesh(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) (string, error) {
	segments := strings.Split(resourceName, "/")
	var resourceType, name string
	if len(segments) > 2 {
		return "", fmt.Errorf("invalid resource name: %s", resourceName)
	} else if len(segments) == 2 {
		resourceType = segments[0]
		name = segments[1]
	} else {
		resourceType = "deployment"
		name = resourceName
	}

	var svcName string
	switch resourceType {
	case "deploy":
	case "deployment":
		app, err := k.GetDeployment(ctx, name, opts.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to find deployment '%s'", name)
			return "", err
		}
		svc, err := getServiceByDeployment(ctx, k, app, opts)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to find deployment '%s'", name)
			return "", err
		}
		svcName = svc.Name
	case "svc":
	case "service":
		svcName = name
	default:
		return "", fmt.Errorf("invalid resource type: %s", resourceType)
	}
	return svcName, nil
}

func lockAndFetchService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string, times int) (*coreV1.Service, error) {
	if times > 10 {
		log.Warn().Msgf("Unable to obtain auto mesh lock, please try again later.")
		return nil, fmt.Errorf("failed to obtain auto meth lock of service %s", serviceName)
	}
	svc, err := k.GetService(ctx, serviceName, namespace)
	if err != nil {
		return nil, err
	}

	if _, ok := svc.Annotations[common.KtLock]; ok {
		log.Info().Msgf("Another user is meshing service %s, waiting for lock ...", serviceName)
		time.Sleep(3 * time.Second)
		return lockAndFetchService(ctx, k, serviceName, namespace, times + 1)
	} else {
		util.MapPut(svc.Annotations, common.KtLock, util.GetTimestamp())
		if svc, err = k.UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock service %s", serviceName)
			return lockAndFetchService(ctx, k, serviceName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Service %s locked for auto mesh", serviceName)
	return svc, nil
}

func unlockService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string) {
	svc, err := k.GetService(ctx, serviceName, namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get service %s for unlock", serviceName)
		return
	}
	if _, ok := svc.Annotations[common.KtLock]; ok {
		delete(svc.Annotations, common.KtLock)
		if _, err2 := k.UpdateService(ctx, svc); err2 != nil {
			log.Warn().Err(err2).Msgf("Failed to unlock service %s", serviceName)
		}
	} else {
		log.Info().Msgf("Service %s doesn't have lock", serviceName)
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
