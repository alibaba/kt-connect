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
	CreatePod(ctx context.Context, metaAndSpec *PodMetaAndSpec, options *options.DaemonOptions) error
	CreateShadowPod(ctx context.Context, metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) error
	GetPod(ctx context.Context, name string, namespace string) (*coreV1.Pod, error)
	GetPods(ctx context.Context, labels map[string]string, namespace string) (pods *coreV1.PodList, err error)
	UpdatePod(ctx context.Context, pod *coreV1.Pod) (*coreV1.Pod, error)
	WaitPodReady(name, namespace string) (pod *coreV1.Pod, err error)
	IncreaseRef(ctx context.Context, name ,namespace string) error
	DecreaseRef(ctx context.Context, name, namespace string) (cleanup bool, err error)
	AddEphemeralContainer(ctx context.Context, containerName, podName string, options *options.DaemonOptions, envs map[string]string) (sshcm string, err error)
	RemoveEphemeralContainer(ctx context.Context, containerName, podName string, namespace string) (err error)
	ExecInPod(containerName, podName, namespace string, opts options.RuntimeOptions, cmd ...string) (string, string, error)
	RemovePod(ctx context.Context, name, namespace string) (err error)

	GetDeployment(ctx context.Context, name string, namespace string) (*appV1.Deployment, error)
	UpdateDeployment(ctx context.Context, deployment *appV1.Deployment) (*appV1.Deployment, error)
	ScaleTo(ctx context.Context, deployment, namespace string, replicas *int32) (err error)

	CreateService(ctx context.Context, name, namespace string, external bool, port map[int]int, labels map[string]string) (*coreV1.Service, error)
	GetService(ctx context.Context, name, namespace string) (*coreV1.Service, error)
	GetServices(ctx context.Context, matchLabels map[string]string, namespace string) ([]coreV1.Service, error)
	GetServiceHosts(ctx context.Context, namespace string) (hosts map[string]string)
	RemoveService(ctx context.Context, name, namespace string) (err error)

	CreateConfigMapWithSshKey(ctx context.Context, labels map[string]string, sshcm string, namespace string,
		generator *util.SSHGenerator) (configMap *coreV1.ConfigMap, err error)
	GetConfigMap(ctx context.Context, name, namespace string) (*coreV1.ConfigMap, error)
	RemoveConfigMap(ctx context.Context, name, namespace string) (err error)

	ClusterCidrs(ctx context.Context, namespace string, connectOptions *options.ConnectOptions) (cidrs []string, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	KubeConfig string
	Clientset kubernetes.Interface
}
