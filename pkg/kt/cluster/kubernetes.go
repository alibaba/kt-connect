package cluster

import (
	"bytes"
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
	"k8s.io/apimachinery/pkg/fields"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"strings"
	"time"
)

// PodMetaAndSpec ...
type PodMetaAndSpec struct {
	Meta  *ResourceMeta
	Image string
	Envs  map[string]string
}

// SvcMetaAndSpec ...
type SvcMetaAndSpec struct {
	Meta      *ResourceMeta
	External  bool
	Ports     map[int]int
	Selectors map[string]string
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

	log.Info().Msgf("Scaling deployment %s to %d", deployment.Name, *replicas)
	deployment.Spec.Replicas = replicas

	if _, err = k.UpdateDeployment(ctx, deployment); err != nil {
		log.Error().Err(err).Msgf("Failed to scale deployment %s", deployment.Name)
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
func (k *Kubernetes) GetServices(ctx context.Context, matchLabels map[string]string, namespace string) ([]coreV1.Service, error) {
	var matchedSvcs []coreV1.Service
	svcList, err := k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, svc := range svcList.Items {
		if util.MapContains(svc.Spec.Selector, matchLabels) {
			matchedSvcs = append(matchedSvcs, svc)
		}
	}
	return matchedSvcs, nil
}

// GetConfigMap get configmap
func (k *Kubernetes) GetConfigMap(ctx context.Context, name, namespace string) (*coreV1.ConfigMap, error) {
	return k.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetConfigMapsByLabel get deployments by label
func (k *Kubernetes) GetConfigMapsByLabel(ctx context.Context, labels map[string]string, namespace string) (pods *coreV1.ConfigMapList, err error) {
	return k.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// GetDeployment ...
func (k *Kubernetes) GetDeployment(ctx context.Context, name string, namespace string) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetDeploymentsByLabel get deployments by label
func (k *Kubernetes) GetDeploymentsByLabel(ctx context.Context, labels map[string]string, namespace string) (pods *appV1.DeploymentList, err error) {
	return k.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// GetPod ...
func (k *Kubernetes) GetPod(ctx context.Context, name string, namespace string) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPodsByLabel get pods by label
func (k *Kubernetes) GetPodsByLabel(ctx context.Context, labels map[string]string, namespace string) (*coreV1.PodList, error) {
	return k.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// AddEphemeralContainer add ephemeral container to specified pod
func (k *Kubernetes) AddEphemeralContainer(ctx context.Context, containerName string, name string,
	options *options.DaemonOptions, envs map[string]string) (string, error) {
	pod, err := k.GetPod(ctx, name, options.Namespace)
	if err != nil {
		return "", err
	}

	privateKeyPath := util.PrivateKeyPath(name)
	generator, err := util.Generate(privateKeyPath)
	if err != nil {
		return "", err
	}
	configMap, err2 := k.CreateConfigMapWithSshKey(ctx, map[string]string{}, name, options.Namespace, generator)

	if err2 != nil {
		return "", errors.New("Found shadow pod but no configMap. Please delete the pod " + pod.Name)
	}

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SshAuthPrivateKey]))

	authKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[common.SshAuthKey]))
	privateKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[common.SshAuthPrivateKey]))

	ec := coreV1.EphemeralContainer{
		EphemeralContainerCommon: coreV1.EphemeralContainerCommon{
			Name:  containerName,
			Image: options.Image,
			Env: []coreV1.EnvVar{
				{Name: common.SshAuthKey, Value: authKey},
				{Name: common.SshAuthPrivateKey, Value: privateKey},
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
	return privateKeyPath, err
}

// RemoveEphemeralContainer remove ephemeral container from specified pod
func (k *Kubernetes) RemoveEphemeralContainer(ctx context.Context, _, podName string, namespace string) (err error) {
	// TODO: implement container removal
	return k.RemovePod(ctx, podName, namespace)
}

func (k *Kubernetes) ExecInPod(containerName, podName, namespace string, opts options.RuntimeOptions, cmd ...string) (string, string, error) {
	req := k.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)
	req.VersionedParams(&coreV1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	log.Debug().Msgf("Execute command %v in %s:%s", cmd, podName, containerName)
	err := execute("POST", req.URL(), opts.RestConfig, nil, &stdout, &stderr, false)
	stdoutMsg := util.RemoveColor(strings.TrimSpace(stdout.String()))
	stderrMsg := util.RemoveColor(strings.TrimSpace(stderr.String()))
	rawErrMsg := util.ExtractErrorMessage(stderrMsg)
	if err == nil && rawErrMsg != "" {
		err = fmt.Errorf(rawErrMsg)
	}
	return stdoutMsg, stderrMsg, err
}

func (k *Kubernetes) CreateConfigMapWithSshKey(ctx context.Context, labels map[string]string, sshcm string, namespace string,
	generator *util.SSHGenerator) (configMap *coreV1.ConfigMap, err error) {

	annotations := map[string]string{common.KtLastHeartBeat: util.GetTimestamp()}
	labels[common.KtName] = sshcm
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
			common.SshAuthKey:        string(generator.PublicKey),
			common.SshAuthPrivateKey: string(generator.PrivateKey),
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
			MountPath: fmt.Sprintf("/root/%s", common.SshAuthKey),
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
func (k *Kubernetes) CreateService(ctx context.Context, metaAndSpec *SvcMetaAndSpec) (*coreV1.Service, error) {
	cli := k.Clientset.CoreV1().Services(metaAndSpec.Meta.Namespace)
	util.SetupServiceHeartBeat(ctx, cli, metaAndSpec.Meta.Name)
	svc := createService(metaAndSpec)
	return cli.Create(ctx, svc, metav1.CreateOptions{})
}

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(ctx context.Context, namespace string, opt *options.ConnectOptions) (cidrs []string, err error) {
	if !opt.DisablePodIp {
		cidrs, err = getPodCidrs(ctx, k.Clientset, namespace, opt.CIDRs)
		if err != nil {
			return
		}
	}
	log.Debug().Msgf("Pod CIDR is %v", cidrs)

	serviceCidr, err := getServiceCidr(ctx, k.Clientset, namespace)
	if err != nil {
		return
	}
	log.Debug().Msgf("Service CIDR is %v", serviceCidr)

	cidrs = append(cidrs, serviceCidr...)
	return
}

// GetServiceHosts get service dns map
func (k *Kubernetes) GetServiceHosts(ctx context.Context, namespace string) (hosts map[string]string) {
	services, err := k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	hosts = map[string]string{}
	for _, service := range services.Items {
		hosts[service.Name] = service.Spec.ClusterIP
	}
	return
}

// GetServicesByLabel get services by label
func (k *Kubernetes) GetServicesByLabel(ctx context.Context, labels map[string]string, namespace string) (svcs *coreV1.ServiceList, err error) {
	return k.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// WatchService ...
func (k *Kubernetes) WatchService(name, namespace string, f func(*coreV1.Service)) {
	watchlist := cache.NewListWatchFromClient(
		k.Clientset.CoreV1().RESTClient(),
		string(coreV1.ResourceServices),
		namespace,
		fields.OneTermEqualSelector("metadata.name", name),
	)
	_, controller := cache.NewInformer(
		watchlist,
		&coreV1.Service{},
		0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				f(newObj.(*coreV1.Service))
			},
		},
	)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(1000 * time.Second)
	}
}

// WaitPodReady ...
func (k *Kubernetes) WaitPodReady(name, namespace string) (pod *coreV1.Pod, err error) {
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	podListener, err := clusterWatcher.PodListenerWithNamespace(k.Clientset, namespace, stopSignal)
	if err != nil {
		return
	}
	pod = &coreV1.Pod{}
	podLabels := k8sLabels.NewSelector()
	requirement, err := k8sLabels.NewRequirement(common.KtName, selection.Equals, []string{name})
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
			p := getTargetPod(common.KtName, name, pods)
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
	return k.Clientset.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
}

// UpdateDeployment ...
func (k *Kubernetes) UpdateDeployment(ctx context.Context, deployment *appV1.Deployment) (*appV1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

// UpdateService ...
func (k *Kubernetes) UpdateService(ctx context.Context, svc *coreV1.Service) (*coreV1.Service, error) {
	return k.Clientset.CoreV1().Services(svc.Namespace).Update(ctx, svc, metav1.UpdateOptions{})
}

// IncreaseRef increase pod ref count by 1
func (k *Kubernetes) IncreaseRef(ctx context.Context, name string, namespace string) error {
	pod, err := k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := pod.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[common.KtRefCount])
	if err != nil {
		log.Error().Err(err).Msgf("Failed to parse annotations[%s] of pod %s with value %s",
			common.KtRefCount, name, annotations[common.KtRefCount])
		return err
	}

	pod.Annotations[common.KtRefCount] = strconv.Itoa(count + 1)

	_, err = k.Clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}

// DecreaseRef decrease pod ref count by 1
func (k *Kubernetes) DecreaseRef(ctx context.Context, name string, namespace string) (cleanup bool, err error) {
	pod, err := k.GetPod(ctx, name, namespace)
	if err != nil {
		return
	}
	refCount := pod.Annotations[common.KtRefCount]
	if refCount == "1" {
		cleanup = true
		log.Info().Msgf("Pod %s has only one ref, gonna remove", name)
		err = k.RemovePod(ctx, pod.Name, pod.Namespace)
	} else {
		err = k.decreasePodRefByOne(ctx, refCount, pod)
	}
	return
}

func (k *Kubernetes) decreasePodRefByOne(ctx context.Context, refCount string, pod *coreV1.Pod) (err error) {
	count, err := decreaseRef(refCount)
	if err != nil {
		return
	}
	log.Info().Msgf("Pod %s has %s refs, decrease to %s", pod.Name, refCount, count)
	util.MapPut(pod.Annotations, common.KtRefCount, count)
	_, err = k.UpdatePod(ctx, pod)
	return
}
