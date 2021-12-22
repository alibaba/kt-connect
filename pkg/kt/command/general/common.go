package general

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

func CreateShadowAndInbound(ctx context.Context, k cluster.KubernetesInterface, shadowPodName string,
	labels, annotations map[string]string, options *options.DaemonOptions) error {

	labels[common.ControlBy] = common.KubernetesTool
	envs := make(map[string]string)
	_, podName, credential, err := cluster.GetOrCreateShadow(ctx, k, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}

	// record context data
	options.RuntimeOptions.Shadow = shadowPodName

	if _, err = tunnel.ForwardPodToLocal(options.MeshOptions.Expose, podName, credential.PrivateKeyPath, options); err != nil {
		return err
	}
	return nil
}

func GetServiceByResourceName(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) (string, error) {
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

func LockAndFetchService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string, times int) (*coreV1.Service, error) {
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
		return LockAndFetchService(ctx, k, serviceName, namespace, times + 1)
	} else {
		util.MapPut(svc.Annotations, common.KtLock, util.GetTimestamp())
		if svc, err = k.UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock service %s", serviceName)
			return LockAndFetchService(ctx, k, serviceName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Service %s locked for auto mesh", serviceName)
	return svc, nil
}

func UnlockService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string) {
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

func UpdateServiceSelector(ctx context.Context, k cluster.KubernetesInterface,svc *coreV1.Service, selector map[string]string) error {
	if _, ok := svc.Annotations[common.KtSelector]; !ok {
		if marshaledSelector, err := json.Marshal(svc.Spec.Selector); err != nil {
			log.Error().Err(err).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err
		} else {
			util.MapPut(svc.Annotations, common.KtSelector, string(marshaledSelector))
		}
	}
	svc.Spec.Selector = selector
	if _, err := k.UpdateService(ctx, svc); err != nil {
		return err
	}
	return nil
}

func getServiceByDeployment(ctx context.Context, k cluster.KubernetesInterface, app *appV1.Deployment,
	options *options.DaemonOptions) (*coreV1.Service, error) {
	svcList, err := k.GetServices(ctx, app.Spec.Selector.MatchLabels, options.Namespace)
	if err != nil {
		return nil, err
	} else if len(svcList) == 0 {
		return nil, fmt.Errorf("failed to find service for deployment '%s', with labels '%v'",
			app.Name, app.Spec.Selector.MatchLabels)
	} else if len(svcList) > 1 {
		svcNames := svcList[0].Name
		for i, svc := range svcList {
			if i > 0 {
				svcNames = svcNames + ", " + svc.Name
			}
		}
		log.Warn().Msgf("Found %d services match deployment '%s': %s. First one will be used.",
			len(svcList), app.Name, svcNames)
	}
	svc := svcList[0]
	if strings.HasSuffix(svc.Name, common.OriginServiceSuffix) {
		return k.GetService(ctx, strings.TrimSuffix(svc.Name, common.OriginServiceSuffix), options.Namespace)
	}
	return &svc, nil
}
