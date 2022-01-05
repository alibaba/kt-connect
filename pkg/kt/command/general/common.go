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

func CreateShadowAndInbound(ctx context.Context, k cluster.KubernetesInterface, shadowPodName, portsToExpose string,
	labels, annotations map[string]string, options *options.DaemonOptions) error {

	podLabels := util.MergeMap(labels, map[string]string{common.ControlBy: common.KubernetesTool})
	envs := make(map[string]string)
	_, podName, credential, err := cluster.GetOrCreateShadow(ctx, k, shadowPodName, options, podLabels, annotations, envs)
	if err != nil {
		return err
	}

	if _, err = tunnel.ForwardPodToLocal(portsToExpose, podName, credential.PrivateKeyPath, options); err != nil {
		return err
	}
	return nil
}

func GetServiceByResourceName(ctx context.Context, k cluster.KubernetesInterface, resourceName, namespace string) (*coreV1.Service, error) {
	resourceType, name, err := ParseResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "deploy":
		fallthrough
	case "deployment":
		app, err2 := k.GetDeployment(ctx, name, namespace)
		if err2 != nil {
			return nil, err2
		}
		return getServiceByDeployment(ctx, k, app, namespace)
	case "svc":
		fallthrough
	case "service":
		return k.GetService(ctx, name, namespace)
	default:
		return nil, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

func GetDeploymentByResourceName(ctx context.Context, k cluster.KubernetesInterface, resourceName, namespace string) (*appV1.Deployment, error) {
	resourceType, name, err := ParseResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "deploy":
		fallthrough
	case "deployment":
		return k.GetDeployment(ctx, name, namespace)
	case "svc":
		fallthrough
	case "service":
		svc, err2 := k.GetService(ctx, name, namespace)
		if err2 != nil {
			return nil, err2
		}
		return getDeploymentByService(ctx, k, svc, namespace)
	default:
		return nil, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

func ParseResourceName(resourceName string) (string, string, error) {
	segments := strings.Split(resourceName, "/")
	var resourceType, name string
	if len(segments) > 2 {
		return "", "", fmt.Errorf("invalid resource name: %s", resourceName)
	} else if len(segments) == 2 {
		resourceType = segments[0]
		name = segments[1]
	} else {
		resourceType = "service"
		name = resourceName
	}
	return resourceType, name, nil
}

func LockService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string, times int) error {
	if times > 10 {
		return fmt.Errorf("failed to obtain kt lock of service %s, please try again later", serviceName)
	}
	svc, err := k.GetService(ctx, serviceName, namespace)
	if err != nil {
		return err
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	if _, ok := svc.Annotations[common.KtLock]; ok {
		log.Info().Msgf("Another user is occupying service %s, waiting for lock ...", serviceName)
		time.Sleep(3 * time.Second)
		return LockService(ctx, k, serviceName, namespace, times + 1)
	} else {
		svc.Annotations[common.KtLock] = util.GetTimestamp()
		if svc, err = k.UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock service %s", serviceName)
			return LockService(ctx, k, serviceName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Service %s locked", serviceName)
	return nil
}

func UnlockService(ctx context.Context, k cluster.KubernetesInterface, serviceName, namespace string) {
	svc, err := k.GetService(ctx, serviceName, namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get service %s for unlock", serviceName)
		return
	}
	if _, ok := svc.Annotations[common.KtLock]; ok {
		delete(svc.Annotations, common.KtLock)
		if _, err = k.UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to unlock service %s", serviceName)
		} else {
			log.Info().Msgf("Service %s unlocked", serviceName)
		}
	} else {
		log.Info().Msgf("Service %s doesn't have lock", serviceName)
	}
}

func UpdateServiceSelector(ctx context.Context, k cluster.KubernetesInterface, svcName, namespace string, selector map[string]string) error {
	svc, err := k.GetService(ctx, svcName, namespace)
	if err != nil {
		return err
	}

	// if KtSelector annotation already exist, fetch current value
	// otherwise you are the first exchange/mesh user to this service, record original selector
	var marshaledSelector string
	if svc.Annotations != nil && svc.Annotations[common.KtSelector] != "" {
		marshaledSelector = svc.Annotations[common.KtSelector]
	} else {
		rawSelector, err2 := json.Marshal(svc.Spec.Selector)
		if err2 != nil {
			log.Error().Err(err2).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err2
		}
		marshaledSelector = string(rawSelector)
		if svc.Annotations == nil {
			util.MapPut(svc.Annotations, common.KtSelector, marshaledSelector)
		} else if _, ok := svc.Annotations[common.KtSelector]; !ok {
			svc.Annotations[common.KtSelector] = marshaledSelector
		}
	}

	if isServiceChanged(svc, selector, marshaledSelector) {
		svc.Spec.Selector = selector
		if _, err = k.UpdateService(ctx, svc); err != nil {
			return err
		}
	}

	go k.WatchService(svcName, namespace, func(newSvc *coreV1.Service) {
		if !isServiceChanged(newSvc, selector, marshaledSelector) {
			return
		}
		log.Debug().Msgf("Change in service %s detected", svcName)
		time.Sleep(util.RandomSeconds(1, 10))
		if svc, err = k.GetService(ctx, svcName, namespace); err == nil {
			if isServiceChanged(svc, selector, marshaledSelector) {
				svc.Spec.Selector = selector
				util.MapPut(svc.Annotations, common.KtSelector, marshaledSelector)
				if _, err = k.UpdateService(ctx, svc); err != nil {
					log.Error().Err(err).Msgf("Failed to recover service %s", svcName)
				} else {
					log.Info().Msgf("Service %s recovered", svcName)
				}
			}
		}
	})
	return nil
}

func isServiceChanged(svc *coreV1.Service, selector map[string]string, marshaledSelector string) bool {
	return !util.MapEquals(svc.Spec.Selector, selector) || svc.Annotations == nil || svc.Annotations[common.KtSelector] != marshaledSelector
}

func getServiceByDeployment(ctx context.Context, k cluster.KubernetesInterface, app *appV1.Deployment,
	namespace string) (*coreV1.Service, error) {
	svcList, err := k.GetServicesBySelector(ctx, app.Spec.Selector.MatchLabels, namespace)
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
		return k.GetService(ctx, strings.TrimSuffix(svc.Name, common.OriginServiceSuffix), namespace)
	}
	return &svc, nil
}

func getDeploymentByService(ctx context.Context, k cluster.KubernetesInterface, svc *coreV1.Service, namespace string) (*appV1.Deployment, error) {
	apps, err := k.GetAllDeploymentInNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	for _, app := range apps.Items {
		if util.MapContains(svc.Spec.Selector, app.Spec.Template.Labels) {
			log.Info().Msgf("Using first matched deployment '%s'", app.Name)
			return &app, nil
		}
	}
	return nil, fmt.Errorf("failed to find deployment for service '%s', with selector '%v'", svc.Name, svc.Spec.Selector)
}
