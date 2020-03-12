package cluster

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
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
	RemoveDeployment(name, namespace string) (err error)
	RemoveConfigMap(name, namespace string) (err error)
	RemoveService(name, namespace string) (err error)
	Deployment(name, namespace string) (deployment *appV1.Deployment, err error)
	Scale(deployment *appV1.Deployment, replicas *int32) (err error)
	ScaleTo(deployment, namespace string, replicas *int32) (err error)
	ServiceHosts(namespace string) (hosts map[string]string)
	ClusterCrids(podCIDR string) (cidrs []string, err error)
	CreateShadow(name, namespace, image string, labels map[string]string, debug bool) (podIP, podName, sshcm string, credential *util.SSHCredential, err error)
	CreateService(name, namespace string, port int, labels map[string]string) (*coreV1.Service, error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	KubeConfig string
	// TODO: should remove
	Clientset kubernetes.Interface
	// TODO: should remove
	ServiceListener v1.ServiceLister
}
