package general

import (
	"encoding/json"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"
)

func CreateShadowAndInbound(shadowPodName, portsToExpose string, labels, annotations map[string]string, portNameDict map[int]string) error {

	envs := make(map[string]string)
	_, podName, privateKeyPath, err := cluster.Ins().GetOrCreateShadow(shadowPodName, labels, annotations, envs, portsToExpose, portNameDict)
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
			if k8sErrors.IsNotFound(err2) {
				return nil, fmt.Errorf("deployment '%s' is not found in namespace %s", name, namespace)
			}
			return nil, err2
		}
		return getServiceByDeployment(app, namespace)
	case "svc":
		fallthrough
	case "service":
		svc, err2 := cluster.Ins().GetService(name, namespace)
		if err2 != nil && k8sErrors.IsNotFound(err2) {
			return nil, fmt.Errorf("service '%s' is not found in namespace %s", name, namespace)
		}
		return svc, err2
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
		app, err2 := cluster.Ins().GetDeployment(name, namespace)
		if err2 != nil && k8sErrors.IsNotFound(err2) {
			return nil, fmt.Errorf("deployment '%s' is not found in namespace %s", name, namespace)
		}
		return app, err2
	case "svc":
		fallthrough
	case "service":
		svc, err2 := cluster.Ins().GetService(name, namespace)
		if err2 != nil {
			if k8sErrors.IsNotFound(err2) {
				return nil, fmt.Errorf("service '%s' is not found in namespace %s", name, namespace)
			}
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
	} else if svc.Spec.Selector[util.KtRole] != "" {
		// service has no kt-selector annotation, but already point to a shadow or router pod
		return fmt.Errorf("exchange or mesh service selecting kt pods is not allow")
	} else {
		rawSelector, err2 := json.Marshal(svc.Spec.Selector)
		if err2 != nil {
			log.Error().Err(err2).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err2
		}
		marshaledSelector = string(rawSelector)
		if svc.Annotations == nil || svc.Annotations[util.KtSelector] == "" {
			svc.Annotations = util.MapPut(svc.Annotations, util.KtSelector, marshaledSelector)
		}
	}

	if isServiceChanged(svc, selector, marshaledSelector) {
		svc.Spec.Selector = selector
		if _, err = cluster.Ins().UpdateService(svc); err != nil {
			return err
		}
	}

	go cluster.Ins().WatchService(svcName, namespace, nil, nil, func(newSvc *coreV1.Service) {
		if pods, err2 := cluster.Ins().GetPodsByLabel(selector, namespace); err2 != nil || len(pods.Items) == 0 {
			log.Warn().Msgf("Router pod has gone")
			return
		} else if pods.Items[0].DeletionTimestamp != nil {
			log.Warn().Msgf("Router pod is terminating")
			return
		} else if len(pods.Items) > 1 {
			log.Warn().Msgf("More than one router pod selected")
		}
		if !isServiceChanged(newSvc, selector, marshaledSelector) {
			return
		}
		log.Debug().Msgf("Change in service %s detected", svcName)
		// delay and double check to avoid multiple clients conflict
		time.Sleep(util.RandomSeconds(1, 10))
		if svc, err = cluster.Ins().GetService(svcName, namespace); err == nil {
			if isServiceChanged(svc, selector, marshaledSelector) {
				svc.Spec.Selector = selector
				svc.Annotations = util.MapPut(svc.Annotations, util.KtSelector, marshaledSelector)
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

func GetTargetPorts(svc *coreV1.Service) map[int]string {
	var pod *coreV1.Pod = nil
	svcPorts := svc.Spec.Ports
	targetPorts := map[int]string{}
	for _, p := range svcPorts {
		if p.TargetPort.Type == intstr.Int {
			targetPorts[p.TargetPort.IntValue()] = fmt.Sprintf("kt-%d", p.TargetPort.IntValue())
		} else {
			if pod == nil {
				pods, err := cluster.Ins().GetPodsByLabel(svc.Spec.Selector, opt.Get().Global.Namespace)
				if err != nil || len(pods.Items) == 0 {
					return map[int]string{}
				}
				pod = &pods.Items[0]
			}
			for _, c := range pod.Spec.Containers {
				for _, cp := range c.Ports {
					if cp.Name == p.TargetPort.String() {
						targetPorts[int(cp.ContainerPort)] = cp.Name
						continue
					}
				}
			}
		}
	}
	return targetPorts
}

func isServiceChanged(svc *coreV1.Service, selector map[string]string, marshaledSelector string) bool {
	return !util.MapEquals(svc.Spec.Selector, selector) || svc.Annotations == nil || svc.Annotations[util.KtSelector] != marshaledSelector
}

func getServiceByDeployment(app *appV1.Deployment, namespace string) (*coreV1.Service, error) {
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

func GetOccupiedUser(labels map[string]string) string {
	podList, err3 := cluster.Ins().GetPodsByLabel(labels, opt.Get().Global.Namespace)
	if err3 == nil && len(podList.Items) > 0 && podList.Items[0].Annotations != nil && podList.Items[0].Annotations[util.KtUser] != "" {
		return " (" + podList.Items[0].Annotations[util.KtUser] + ")"
	}
	appList, err := cluster.Ins().GetDeploymentsByLabel(labels, opt.Get().Global.Namespace)
	if err == nil && len(appList.Items) > 0 && appList.Items[0].Annotations != nil && appList.Items[0].Annotations[util.KtUser] != "" {
		return " (" + appList.Items[0].Annotations[util.KtUser] + ")"
	}
	return ""
}
