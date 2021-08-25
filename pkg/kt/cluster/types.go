package cluster

import (
	"context"
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
	RemoveDeployment(ctx context.Context, name, namespace string) (err error)
	RemoveConfigMap(ctx context.Context, name, namespace string) (err error)
	RemoveService(ctx context.Context, name, namespace string) (err error)
	Deployment(ctx context.Context, name, namespace string) (deployment *appV1.Deployment, err error)
	Scale(ctx context.Context, deployment *appV1.Deployment, replicas *int32) (err error)
	ScaleTo(ctx context.Context, deployment, namespace string, replicas *int32) (err error)
	ServiceHosts(ctx context.Context, namespace string) (hosts map[string]string)
	ClusterCidrs(ctx context.Context, namespace string, connectOptions *options.ConnectOptions) (cidrs []string, err error)
	GetOrCreateShadow(ctx context.Context, name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (podIP, podName, sshcm string, credential *util.SSHCredential, err error)
	GetAllExistingShadowDeployments(ctx context.Context, namespace string) (list []appV1.Deployment, err error)
	CreateService(ctx context.Context, name, namespace string, external bool, port int, labels map[string]string) (*coreV1.Service, error)
	GetDeployment(ctx context.Context, name string, namespace string) (*appV1.Deployment, error)
	UpdateDeployment(ctx context.Context, namespace string, deployment *appV1.Deployment) (*appV1.Deployment, error)
	DecreaseRef(ctx context.Context, namespace string, deployment string) (cleanup bool, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	KubeConfig string
	Clientset  kubernetes.Interface
}
