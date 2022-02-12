package general

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

func CreateShadowAndInbound(shadowPodName, portsToExpose string,
	labels, annotations map[string]string) error {

	podLabels := util.MergeMap(labels, map[string]string{util.ControlBy: util.KubernetesToolkit})
	envs := make(map[string]string)
	_, podName, privateKeyPath, err := cluster.Ins().GetOrCreateShadow(shadowPodName, podLabels, annotations, envs)
	if err != nil {
		return err
	}

	if _, err = transmission.ForwardPodToLocal(portsToExpose, podName, privateKeyPath); err != nil {
		return err
	}
	return nil
}

func GetServiceByResourceName(resourceName, namespace string) (*coreV1.Service, error) {
	resourceType, name, err := ParseResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "deploy":
		fallthrough
	case "deployment":
		app, err2 := cluster.Ins().GetDeployment(name, namespace)
		if err2 != nil {
			return nil, err2
		}
		return getServiceByDeployment(app, namespace)
	case "svc":
		fallthrough
	case "service":
		return cluster.Ins().GetService(name, namespace)
	default:
		return nil, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

func GetDeploymentByResourceName(resourceName, namespace string) (*appV1.Deployment, error) {
	resourceType, name, err := ParseResourceName(resourceName)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "deploy":
		fallthrough
	case "deployment":
		return cluster.Ins().GetDeployment(name, namespace)
	case "svc":
		fallthrough
	case "service":
		svc, err2 := cluster.Ins().GetService(name, namespace)
		if err2 != nil {
			return nil, err2
		}
		return getDeploymentByService(svc, namespace)
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

func UpdateServiceSelector(svcName, namespace string, selector map[string]string) error {
	svc, err := cluster.Ins().GetService(svcName, namespace)
	if err != nil {
		return err
	}

	// if KtSelector annotation already exist, fetch current value
	// otherwise you are the first exchange/mesh user to this service, record original selector
	var marshaledSelector string
	if svc.Annotations != nil && svc.Annotations[util.KtSelector] != "" {
		marshaledSelector = svc.Annotations[util.KtSelector]
	} else {
		rawSelector, err2 := json.Marshal(svc.Spec.Selector)
		if err2 != nil {
			log.Error().Err(err2).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err2
		}
		marshaledSelector = string(rawSelector)
		if svc.Annotations == nil {
			util.MapPut(svc.Annotations, util.KtSelector, marshaledSelector)
		} else if svc.Annotations[util.KtSelector] == "" {
			svc.Annotations[util.KtSelector] = marshaledSelector
		}
	}

	if isServiceChanged(svc, selector, marshaledSelector) {
		svc.Spec.Selector = selector
		if _, err = cluster.Ins().UpdateService(svc); err != nil {
			return err
		}
	}

	go cluster.Ins().WatchService(svcName, namespace, nil, nil, func(newSvc *coreV1.Service) {
		if !isServiceChanged(newSvc, selector, marshaledSelector) {
			return
		}
		log.Debug().Msgf("Change in service %s detected", svcName)
		time.Sleep(util.RandomSeconds(1, 10))
		if svc, err = cluster.Ins().GetService(svcName, namespace); err == nil {
			if isServiceChanged(svc, selector, marshaledSelector) {
				svc.Spec.Selector = selector
				util.MapPut(svc.Annotations, util.KtSelector, marshaledSelector)
				if _, err = cluster.Ins().UpdateService(svc); err != nil {
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
	return !util.MapEquals(svc.Spec.Selector, selector) || svc.Annotations == nil || svc.Annotations[util.KtSelector] != marshaledSelector
}

func getServiceByDeployment(app *appV1.Deployment,
	namespace string) (*coreV1.Service, error) {
	svcList, err := cluster.Ins().GetServicesBySelector(app.Spec.Selector.MatchLabels, namespace)
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
	if strings.HasSuffix(svc.Name, util.StuntmanServiceSuffix) {
		return cluster.Ins().GetService(strings.TrimSuffix(svc.Name, util.StuntmanServiceSuffix), namespace)
	}
	return &svc, nil
}

func getDeploymentByService(svc *coreV1.Service, namespace string) (*appV1.Deployment, error) {
	apps, err := cluster.Ins().GetAllDeploymentInNamespace(namespace)
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
