package cluster

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateFromClientSet kubernetes instance
func CreateFromClientSet(clientSet kubernetes.Interface) (kubernetes KubernetesInterface, err error) {
	return &Kubernetes{
		Clientset: clientSet,
	}, nil
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
	ClusterCidrs(namespace string, connectOptions *options.ConnectOptions) (cidrs []string, err error)
	GetOrCreateShadow(name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (podIP, podName, sshcm string, credential *util.SSHCredential, err error)
	GetAllExistingShadowDeployments(namespace string) (list []appV1.Deployment, err error)
	CreateService(name, namespace string, external bool, port int, labels map[string]string) (*coreV1.Service, error)
	GetDeployment(name string, namespace string) (*appV1.Deployment, error)
	UpdateDeployment(namespace string, deployment *appV1.Deployment) (*appV1.Deployment, error)
	DecreaseRef(namespace string, deployment string) (cleanup bool, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	KubeConfig string
	Clientset  kubernetes.Interface
}
