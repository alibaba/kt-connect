package cluster

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesClient(kubeConfig string) (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	return
}

func createService(metaAndSpec *SvcMetaAndSpec) *coreV1.Service {
	var servicePorts []coreV1.ServicePort
	util.MapPut(metaAndSpec.Meta.Annotations, util.KtLastHeartBeat, util.GetTimestamp())
	util.MapPut(metaAndSpec.Meta.Labels, util.ControlBy, util.KubernetesToolkit)

	for srcPort, targetPort := range metaAndSpec.Ports {
		servicePorts = append(servicePorts, coreV1.ServicePort{
			Name:       fmt.Sprintf("kt-%d", srcPort),
			Port:       int32(srcPort),
			TargetPort: intstr.FromInt(targetPort),
		})
	}

	service := &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaAndSpec.Meta.Name,
			Namespace:   metaAndSpec.Meta.Namespace,
			Labels:      metaAndSpec.Meta.Labels,
			Annotations: metaAndSpec.Meta.Annotations,
		},
		Spec: coreV1.ServiceSpec{
			Selector: metaAndSpec.Selectors,
			Type:     coreV1.ServiceTypeClusterIP,
			Ports:    servicePorts,
		},
	}
	if metaAndSpec.External {
		service.Spec.Type = coreV1.ServiceTypeLoadBalancer
	}
	return service
}

func createPod(metaAndSpec *PodMetaAndSpec) *coreV1.Pod {
	var args []string
	namespace := metaAndSpec.Meta.Namespace
	name := metaAndSpec.Meta.Name
	labels := metaAndSpec.Meta.Labels
	annotations := metaAndSpec.Meta.Annotations
	annotations[util.KtRefCount] = "1"
	annotations[util.KtLastHeartBeat] = util.GetTimestamp()
	ports := metaAndSpec.Ports
	image := metaAndSpec.Image
	envs := metaAndSpec.Envs

	pod := &coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: coreV1.PodSpec{
			ServiceAccountName: opt.Get().ServiceAccount,
			Containers: []coreV1.Container{
				createContainer(image, args, envs, ports),
			},
		},
	}

	if opt.Get().ImagePullSecret != "" {
		addImagePullSecret(pod, opt.Get().ImagePullSecret)
	}

	if opt.Get().NodeSelector != "" {
		pod.Spec.NodeSelector = util.String2Map(opt.Get().NodeSelector)
	}

	return pod
}

func createContainer(image string, args []string, envs map[string]string, ports []int) coreV1.Container {
	var envVar []coreV1.EnvVar
	for k, v := range envs {
		envVar = append(envVar, coreV1.EnvVar{Name: k, Value: v})
	}
	var pullPolicy coreV1.PullPolicy
	if opt.Get().AlwaysUpdateShadow {
		pullPolicy = "Always"
	} else {
		pullPolicy = "IfNotPresent"
	}
	container := coreV1.Container{
		Name:            util.DefaultContainer,
		Image:           image,
		ImagePullPolicy: pullPolicy,
		Args:            args,
		Env:             envVar,
		SecurityContext: &coreV1.SecurityContext{
			Capabilities: &coreV1.Capabilities{
				Add: []coreV1.Capability{
					"AUDIT_WRITE",
				},
			},
		},
		Ports: []coreV1.ContainerPort{},
	}
	for _, port := range ports {
		container.Ports = append(container.Ports, coreV1.ContainerPort{
			// TODO: assume port using http protocol, should be replace with user-defined protocol, for istio constraint
			Name: fmt.Sprintf("http-%d", port),
			Protocol: coreV1.ProtocolTCP,
			ContainerPort: int32(port),
		})
	}
	return container
}
