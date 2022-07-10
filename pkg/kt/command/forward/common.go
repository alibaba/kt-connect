package forward

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func RedirectService(serviceName string, localPort, remotePort int) (int, error) {
	podName, podPort, svcPort, err := getPodNameAndPort(serviceName, remotePort, opt.Get().Global.Namespace)
	if err != nil {
		return 0, err
	}
	if localPort <= 0 {
		// local port note provided, use same as remote port
		localPort = svcPort
	}
	gone, err := transmission.SetupPortForwardToLocal(podName, podPort, localPort)
	go func() {
		<-gone
	}()
	return localPort, err
}

func RedirectAddress(remoteAddress string, localPort, remotePort int) error {
	if remotePort <= 0 {
		if localPort <= 0 {
			return fmt.Errorf("port parameter must be specified")
		} else {
			remotePort = localPort
		}
	}
	return fmt.Errorf("redirecting to an arbitrary address havn't been implemented yet")
}

func getPodNameAndPort(serviceName string, remotePort int, namespace string) (string, int, int, error) {
	svc, err := cluster.Ins().GetService(serviceName, namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return "", 0, 0, fmt.Errorf("service '%s' is not found in namespace %s", serviceName, namespace)
		}
		return "", 0, 0, err
	}
	if len(svc.Spec.Ports) == 0 {
		return "", 0, 0, fmt.Errorf("service '%s' has not port available", serviceName)
	}

	if remotePort <= 0 {
		// remote port not provided, try fetch from service
		if len(svc.Spec.Ports) > 1 {
			return "", 0, 0, fmt.Errorf("service '%s' has multiple ports, must specify one", serviceName)
		} else {
			remotePort = int(svc.Spec.Ports[0].Port)
		}
	}

	targetPort := intstr.IntOrString{Type: -1}
	for _, p := range svc.Spec.Ports {
		if int(p.Port) == remotePort {
			targetPort = p.TargetPort
		}
	}
	if targetPort.Type == -1 {
		return "", 0, 0, fmt.Errorf("port %d not available for service %s", remotePort, serviceName)
	}
	pods, err := cluster.Ins().GetPodsByLabel(svc.Spec.Selector, opt.Get().Global.Namespace)
	if err != nil {
		return "", 0, 0, err
	}
	if len(pods.Items) == 0 {
		return "", 0, 0, fmt.Errorf("no pod available for service %s", serviceName)
	}
	podPort := -1
	if targetPort.Type == intstr.Int {
		podPort = int(targetPort.IntVal)
	} else {
	containerLoop:
		for _, c := range pods.Items[0].Spec.Containers {
			for _, p := range c.Ports {
				if p.Name == targetPort.StrVal {
					podPort = int(p.ContainerPort)
					break containerLoop
				}
			}
		}
	}
	if podPort == -1 {
		return "", 0, 0, fmt.Errorf("port %d not fit for any pod of service %s", remotePort, serviceName)
	}
	return pods.Items[0].Name, podPort, remotePort, nil
}
