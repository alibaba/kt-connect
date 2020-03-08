package cluster

import (
	appV1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
)

// Create kubernetes instance
func Create(kubeConfig string) (kubernetes Kubernetes, err error) {
	clientSet, err := GetKubernetesClient(kubeConfig)
	if err != nil {
		return
	}
	kubernetes = Kubernetes{
		Clientset: clientSet,
	}
	return
}

// KubernetesInterface kubernetes interface
type KubernetesInterface interface {
	Deployment(name, namespace string) (deployment *appV1.Deployment, err error)
	Scale(deployment *appV1.Deployment, replicas *int32) (err error)
	ServiceHosts(namespace string) (hosts map[string]string)
	ClusterCrids(podCIDR string) (cidrs []string, err error)
	CreateShadow(name, namespace, image string, labels map[string]string) (podIP, podName string, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	Clientset       *kubernetes.Clientset
	ServiceListener v1.ServiceLister
}
