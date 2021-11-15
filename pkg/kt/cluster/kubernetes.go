package cluster

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	clusterWatcher "github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"strconv"
	"strings"
)

// PodMetaAndSpec ...
type PodMetaAndSpec struct {
	Meta  *ResourceMeta
	Image string
	Envs  map[string]string
}

// ResourceMeta ...
type ResourceMeta struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// SSHkeyMeta ...
type SSHkeyMeta struct {
	SshConfigMapName string
	PrivateKeyPath   string
}

// RemoveService remove service
func (k *Kubernetes) RemoveService(ctx context.Context, name, namespace string) (err error) {
	client := k.Clientset.CoreV1().Services(namespace)
	return client.Delete(ctx, name, metav1.DeleteOptions{})
}

// RemovePod remove pod instances
func (k *Kubernetes) RemovePod(ctx context.Context, name, namespace string) (err error) {
	podsClient := k.Clientset.CoreV1().Pods(namespace)
	deletePolicy := metav1.DeletePropagationBackground
	return podsClient.Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// RemoveConfigMap remove ConfigMap instance
func (k *Kubernetes) RemoveConfigMap(ctx context.Context, name, namespace string) (err error) {
	cli := k.Clientset.CoreV1().ConfigMaps(namespace)
	deletePolicy := metav1.DeletePropagationBackground
	return cli.Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// ScaleTo scale deployment to
func (k *Kubernetes) ScaleTo(ctx context.Context, name, namespace string, replicas *int32) (err error) {
	deployment, err := k.GetDeployment(ctx, name, namespace)
	if err != nil {
		return
	}

	log.Info().Msgf("Scaling deployment %s to %d", deployment.GetObjectMeta().GetName(), *replicas)
	deployment.Spec.Replicas = replicas

	if _, err = k.UpdateDeployment(ctx, deployment); err != nil {
		log.Error().Err(err).Msgf("Fails to scale deployment %s", deployment.GetObjectMeta().GetName())
		return
	}
	log.Info().Msgf("Deployment %s successfully scaled to %d replicas", name, *replicas)
	return
}

// GetService get service
func (k *Kubernetes) GetService(ctx context.Context, name, namespace string) (*coreV1.Service, error) {
	return k.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetServices get pods by label
func (k *Kubernetes) GetServices(ctx context.Context, matchLabels map[string]string, namespace string) (*coreV1.ServiceList, error) {
	return k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: labelApi.SelectorFromSet(matchLabels).String(),
	})
}

// GetConfigMap get configmap
func (k *Kubernetes) GetConfigMap(ctx context.Context, name, namespace string) (*coreV1.ConfigMap, error) {
	return k.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetDeployment ...
func (k *Kubernetes) GetDeployment(ctx context.Context, name string, namespace string) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPod ...
func (k *Kubernetes) GetPod(ctx context.Context, name string, namespace string) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPods get pods by label
func (k *Kubernetes) GetPods(ctx context.Context, labels map[string]string, namespace string) (*coreV1.PodList, error) {
	return k.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// AddEphemeralContainer add ephemeral container to specified pod
func (k *Kubernetes) AddEphemeralContainer(ctx context.Context, containerName string, podName string,
	options *options.DaemonOptions, envs map[string]string) (sshcm string, err error) {
	pod, err := k.GetPod(ctx, podName, options.Namespace)
	if err != nil {
		return "", err
	}
	identifier := strings.ToLower(util.RandomString(4))
	sshcm = fmt.Sprintf("kt-%s-public-key-%s", common.ComponentExchange, identifier)

	privateKeyPath := util.PrivateKeyPath("exchangepod", identifier)
	generator, err := util.Generate(privateKeyPath)
	if err != nil {
		return
	}
	configMap, err2 := k.CreateConfigMapWithSshKey(ctx, map[string]string{}, sshcm, options.Namespace, generator)

	if err2 != nil {
		err = errors.New("Found shadow pod but no configMap. Please delete the pod " + pod.Name)
		return
	}

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SSHAuthPrivateKey]))

	authKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[common.SSHAuthKey]))
	privateKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[common.SSHAuthPrivateKey]))

	ec := coreV1.EphemeralContainer{
		EphemeralContainerCommon: coreV1.EphemeralContainerCommon{
			Name:  containerName,
			Image: options.Image,
			Env: []coreV1.EnvVar{
				{Name: common.SSHAuthKey, Value: authKey},
				{Name: common.SSHAuthPrivateKey, Value: privateKey},
			},
			SecurityContext: &coreV1.SecurityContext{
				Capabilities: &coreV1.Capabilities{Add: []coreV1.Capability{"NET_ADMIN"}},
			},
		},
	}

	for k, v := range envs {
		ec.Env = append(ec.Env, coreV1.EnvVar{Name: k, Value: v})
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ec)

	pod, err = k.Clientset.CoreV1().Pods(pod.Namespace).UpdateEphemeralContainers(ctx, pod.Name, pod, metav1.UpdateOptions{})
	return sshcm, err
}

// RemoveEphemeralContainer remove ephemeral container from specified pod
func (k *Kubernetes) RemoveEphemeralContainer(ctx context.Context, containerName, podName string, namespace string) (err error) {
	// TODO: implement container removal
	return k.RemovePod(ctx, podName, namespace)
}

func (k *Kubernetes) CreateConfigMapWithSshKey(ctx context.Context, labels map[string]string, sshcm string, namespace string,
	generator *util.SSHGenerator) (configMap *coreV1.ConfigMap, err error) {

	annotations := map[string]string{common.KTLastHeartBeat: util.GetTimestamp()}
	labels[common.KTName] = sshcm
	cli := k.Clientset.CoreV1().ConfigMaps(namespace)
	util.SetupConfigMapHeartBeat(ctx, cli, sshcm)

	return cli.Create(ctx, &coreV1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sshcm,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Data: map[string]string{
			common.SSHAuthKey:        string(generator.PublicKey),
			common.SSHAuthPrivateKey: string(generator.PrivateKey),
		},
	}, metav1.CreateOptions{})
}

// CreateShadowPod create shadow pod
func (k *Kubernetes) CreateShadowPod(ctx context.Context, metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) error {
	cli := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace)
	util.SetupPodHeartBeat(ctx, cli, metaAndSpec.Meta.Name)
	pod := createPod(metaAndSpec, options)
	pod.Spec.Containers[0].VolumeMounts = []coreV1.VolumeMount{
		{
			Name:      "ssh-public-key",
			MountPath: fmt.Sprintf("/root/%s", common.SSHAuthKey),
		},
	}
	pod.Spec.Volumes = []coreV1.Volume{
		getSSHVolume(sshcm),
	}
	if options.ConnectOptions != nil && options.ConnectOptions.Method == common.ConnectMethodTun {
		addTunHostPath(pod)
	}
	if _, err := cli.Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

// CreatePod create pod
func (k *Kubernetes) CreatePod(ctx context.Context, metaAndSpec *PodMetaAndSpec, options *options.DaemonOptions) error {
	cli := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace)
	util.SetupPodHeartBeat(ctx, cli, metaAndSpec.Meta.Name)
	pod := createPod(metaAndSpec, options)
	if _, err := cli.Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

// CreateService create kubernetes service
func (k *Kubernetes) CreateService(ctx context.Context, name, namespace string, external bool, port int, labels map[string]string) (*coreV1.Service, error) {
	cli := k.Clientset.CoreV1().Services(namespace)
	util.SetupServiceHeartBeat(ctx, cli, name)
	svc := createService(name, namespace, labels, external, port)
	return cli.Create(ctx, svc, metav1.CreateOptions{})
}

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(ctx context.Context, namespace string, opt *options.ConnectOptions) (cidrs []string, err error) {
	serviceList, err := fetchServiceList(ctx, k, namespace)
	if err != nil {
		return
	}

	if !opt.DisablePodIp {
		cidrs, err = getPodCidrs(ctx, k.Clientset, opt.CIDRs)
		if err != nil {
			return
		}
	}

	services := serviceList.Items
	serviceCidr, err := getServiceCidr(services)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCidr...)
	return
}

// fetchServiceList try list service at cluster scope. fallback to namespace scope
func fetchServiceList(ctx context.Context, k *Kubernetes, namespace string) (*coreV1.ServiceList, error) {
	serviceList, err := k.Clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	}
	return serviceList, err
}

// GetServiceHosts get service dns map
func (k *Kubernetes) GetServiceHosts(ctx context.Context, namespace string) (hosts map[string]string) {
	services, err := k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	hosts = map[string]string{}
	for _, service := range services.Items {
		hosts[service.ObjectMeta.Name] = service.Spec.ClusterIP
	}
	return
}

func (k *Kubernetes) WaitPodReady(name, namespace string) (pod *coreV1.Pod, err error) {
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	podListener, err := clusterWatcher.PodListenerWithNamespace(k.Clientset, namespace, stopSignal)
	if err != nil {
		return
	}
	pod = &coreV1.Pod{}
	podLabels := k8sLabels.NewSelector()
	requirement, err := k8sLabels.NewRequirement(common.KTName, selection.Equals, []string{name})
	if err != nil {
		return
	}
	podLabels.Add(*requirement)

	pods, err := podListener.Pods(namespace).List(podLabels)
	if err != nil {
		return pod, err
	}

	for {
		hasRunningPod := len(pods) > 0
		var podName string
		if hasRunningPod {
			// podLister do not support FieldSelector
			// https://github.com/kubernetes/client-go/issues/604
			p := getTargetPod(common.KTName, name, pods)
			if p != nil {
				if p.Status.Phase == "Running" {
					pod = p
					log.Info().Msgf("Pod %s is ready", pod.Name)
					break
				}
				podName = p.Name
			}
		}
		wait(podName)
		pods, err = podListener.Pods(namespace).List(podLabels)
		if err != nil {
			return pod, err
		}
	}
	return pod, nil
}

// UpdatePod ...
func (k *Kubernetes) UpdatePod(ctx context.Context, pod *coreV1.Pod) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(pod.GetObjectMeta().GetNamespace()).Update(ctx, pod, metav1.UpdateOptions{})
}

func (k *Kubernetes) UpdateDeployment(ctx context.Context, deployment *appV1.Deployment) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(deployment.GetObjectMeta().GetNamespace()).Update(ctx, deployment, metav1.UpdateOptions{})
}

// IncreaseRef increase pod ref count by 1
func (k *Kubernetes) IncreaseRef(ctx context.Context, name string, namespace string) error {
	pod, err := k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := pod.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[common.KTRefCount])
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse annotations[%s] of pod %s with value %s",
			common.KTRefCount, name, annotations[common.KTRefCount])
		return err
	}

	pod.ObjectMeta.Annotations[common.KTRefCount] = strconv.Itoa(count + 1)

	_, err = k.Clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}

// DecreaseRef decrease pod ref count by 1
func (k *Kubernetes) DecreaseRef(ctx context.Context, name string, namespace string) (cleanup bool, err error) {
	pod, err := k.GetPod(ctx, name, namespace)
	if err != nil {
		return
	}
	refCount := pod.ObjectMeta.Annotations[common.KTRefCount]
	if refCount == "1" {
		cleanup = true
		log.Info().Msgf("Shared shadow has only one ref, delete it")
		err = k.RemovePod(ctx, pod.GetObjectMeta().GetName(), pod.GetObjectMeta().GetNamespace())
		if err != nil {
			return
		}
	} else {
		err2 := k.decreasePodRef(ctx, refCount, pod)
		if err2 != nil {
			err = err2
			return
		}
	}
	return
}

func (k *Kubernetes) decreasePodRef(ctx context.Context, refCount string, pod *coreV1.Pod) (err error) {
	log.Info().Msgf("Shared shadow has more than one ref, decrease the ref")
	count, err := decreaseRef(refCount)
	if err != nil {
		return
	}
	pod.ObjectMeta.Annotations[common.KTRefCount] = count
	_, err = k.UpdatePod(ctx, pod)
	return
}

func decreaseRef(refCount string) (count string, err error) {
	currentCount, err := strconv.Atoi(refCount)
	if err != nil {
		return
	}
	count = strconv.Itoa(currentCount - 1)
	return
}
