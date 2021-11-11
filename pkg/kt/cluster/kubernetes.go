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
	appv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
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
func (k *Kubernetes) ScaleTo(ctx context.Context, deployment, namespace string, replicas *int32) (err error) {
	obj, err := k.Deployment(ctx, deployment, namespace)
	if err != nil {
		return
	}
	return k.Scale(ctx, obj, replicas)
}

// Scale scale deployment to
func (k *Kubernetes) Scale(ctx context.Context, deployment *appv1.Deployment, replicas *int32) (err error) {
	log.Info().Msgf("Scaling deployment %s to %d", deployment.GetObjectMeta().GetName(), *replicas)
	client := k.Clientset.AppsV1().Deployments(deployment.GetObjectMeta().GetNamespace())
	deployment.Spec.Replicas = replicas

	d, err := client.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Error().Msgf("Fails to scale deployment %s: %s", deployment.GetObjectMeta().GetName(), err.Error())
		return
	}
	log.Info().Msgf("Deployment %s successfully scaled to %d replicas", d.Name, *d.Spec.Replicas)
	return
}

// Deployment get deployment
func (k *Kubernetes) Deployment(ctx context.Context, name, namespace string) (*appv1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Service get service
func (k *Kubernetes) Service(ctx context.Context, name, namespace string) (*coreV1.Service, error) {
	return k.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Pod get pod
func (k *Kubernetes) Pod(ctx context.Context, name, namespace string) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Pods get pods
func (k *Kubernetes) Pods(ctx context.Context, labels map[string]string, namespace string) (*coreV1.PodList, error) {
	return k.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// AddEphemeralContainer add ephemeral container to specified pod
func (k *Kubernetes) AddEphemeralContainer(ctx context.Context, containerName string, podName string,
	options *options.DaemonOptions, envs map[string]string) (sshcm string, err error) {
	pod, err := k.Pod(ctx, podName, options.Namespace)
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
	configMap, err2 := k.createConfigMap(ctx, map[string]string{}, sshcm, options.Namespace, generator)

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
	return k.RemovePod(ctx, podName, namespace)
}

// GetOrCreateShadow create shadow
func (k *Kubernetes) GetOrCreateShadow(ctx context.Context, name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (
	podIP, podName, sshcm string, credential *util.SSHCredential, err error) {

	component := labels[common.KTComponent]
	identifier := strings.ToLower(util.RandomString(4))
	if options.ConnectOptions.ShareShadow {
		identifier = "shared"
	}
	sshcm = fmt.Sprintf("kt-%s-public-key-%s", component, identifier)
	privateKeyPath := util.PrivateKeyPath(component, identifier)

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.WithLabels) {
		labels[k] = v
	}
	for k, v := range util.String2Map(options.WithAnnotations) {
		annotations[k] = v
	}
	annotations[common.KtUser] = util.GetLocalUserName()

	if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
		pod, generator, err2 := k.tryGetExistingShadowRelatedObjs(ctx, &ResourceMeta{
			Name:        name,
			Namespace:   options.Namespace,
			Labels:      labels,
			Annotations: annotations,
		}, &SSHkeyMeta{
			SshConfigMapName: sshcm,
			PrivateKeyPath:   privateKeyPath,
		})
		if err2 != nil {
			err = err2
			return
		}
		if pod != nil && generator != nil {
			podIP, podName, credential = shadowResult(pod, generator)
			return
		}
	}

	podIP, podName, credential, err = k.createShadow(ctx, &PodMetaAndSpec{
		&ResourceMeta{
			Name:        name,
			Namespace:   options.Namespace,
			Labels:      labels,
			Annotations: annotations,
		}, options.Image, envs,
	}, &SSHkeyMeta{
		SshConfigMapName: sshcm,
		PrivateKeyPath:   privateKeyPath,
	}, options)
	return
}

func (k *Kubernetes) createShadow(ctx context.Context, metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta, options *options.DaemonOptions) (
	podIP string, podName string, credential *util.SSHCredential, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}
	configMap, err2 := k.createConfigMap(ctx, metaAndSpec.Meta.Labels, sshKeyMeta.SshConfigMapName, metaAndSpec.Meta.Namespace, generator)

	if err2 != nil {
		err = err2
		return
	}
	log.Info().Msgf("Successful create config map %v", configMap.ObjectMeta.Name)

	pod, err2 := k.createAndGetPod(ctx, metaAndSpec, sshKeyMeta.SshConfigMapName, options)
	if err2 != nil {
		err = err2
		return
	}
	podIP, podName, credential = shadowResult(pod, generator)
	return
}

// GetAllExistingShadowPods fetch all shadow pods
func (k *Kubernetes) GetAllExistingShadowPods(ctx context.Context, namespace string) ([]coreV1.Pod, error) {
	list, err := k.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: k8sLabels.Set(metav1.LabelSelector{
			MatchLabels: map[string]string{common.ControlBy: common.KubernetesTool},
		}.MatchLabels).String(),
	})
	if list == nil {
		return nil, common.CommandExecError{Reason: "get nil list when querying shadow pods"}
	}
	return list.Items, err
}

func (k *Kubernetes) tryGetExistingShadowRelatedObjs(ctx context.Context, resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (pod *coreV1.Pod, generator *util.SSHGenerator, err error) {
	_, shadowError := k.GetPod(ctx, resourceMeta.Name, resourceMeta.Namespace)
	if shadowError != nil {
		return
	}
	cli := k.Clientset.CoreV1().ConfigMaps(resourceMeta.Namespace)
	configMap, configMapError := cli.Get(ctx, sshKeyMeta.SshConfigMapName, metav1.GetOptions{})

	if configMapError != nil {
		err = errors.New("Found shadow pod but no configMap. Please delete the pod " + resourceMeta.Name)
		return
	}

	generator = util.NewSSHGenerator(configMap.Data[common.SSHAuthPrivateKey], configMap.Data[common.SSHAuthKey], sshKeyMeta.PrivateKeyPath)

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SSHAuthPrivateKey]))
	if err != nil {
		return
	}

	return k.getShadowPod(ctx, resourceMeta, generator)
}

func (k *Kubernetes) getShadowPod(ctx context.Context, resourceMeta *ResourceMeta, generator *util.SSHGenerator) (pod *coreV1.Pod, sshGenerator *util.SSHGenerator, err error) {
	podList, err := k.Clientset.CoreV1().Pods(resourceMeta.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: k8sLabels.Set(metav1.LabelSelector{MatchLabels: resourceMeta.Labels}.MatchLabels).String(),
	})
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shared shadow, reuse it")
		err = increaseRefCount(ctx, resourceMeta.Name, k.Clientset, resourceMeta.Namespace)
		if err != nil {
			return
		}
		return &(podList.Items[0]), generator, nil
	} else if len(podList.Items) > 1 {
		err = errors.New("Found more than one pod with name " + resourceMeta.Name + ", please make sure these is only one in namespace " + resourceMeta.Namespace)
	}
	return
}

func increaseRefCount(ctx context.Context, name string, clientSet kubernetes.Interface, namespace string) error {
	pod, err := clientSet.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := pod.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[common.KTRefCount])
	if err != nil {
		log.Error().Msgf("Failed to parse annotations[%s] of pod %s with value %s",
			common.KTRefCount, name, annotations[common.KTRefCount])
		return err
	}

	pod.ObjectMeta.Annotations[common.KTRefCount] = strconv.Itoa(count + 1)

	_, err = clientSet.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}

func shadowResult(pod *coreV1.Pod, generator *util.SSHGenerator) (string, string, *util.SSHCredential) {
	podIP := pod.Status.PodIP
	podName := pod.GetObjectMeta().GetName()
	credential := util.NewDefaultSSHCredential()
	credential.PrivateKeyPath = generator.PrivateKeyPath
	return podIP, podName, credential
}

func (k *Kubernetes) createAndGetPod(ctx context.Context, metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) (*coreV1.Pod, error) {
	localIPAddress := util.GetOutboundIP()
	log.Debug().Msgf("Client address %s", localIPAddress)
	resourceMeta := metaAndSpec.Meta
	resourceMeta.Labels[common.KTRemoteAddress] = localIPAddress
	resourceMeta.Labels[common.KTName] = resourceMeta.Name
	cli := k.Clientset.CoreV1().Pods(resourceMeta.Namespace)
	util.SetupPodHeartBeat(ctx, cli, resourceMeta.Name)

	pod := createPod(metaAndSpec, sshcm, options)
	result, err := cli.Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("Deploy shadow pod %s in namespace %s", result.GetObjectMeta().GetName(), resourceMeta.Namespace)

	return waitPodReadyUsingInformer(resourceMeta.Namespace, resourceMeta.Name, k.Clientset)
}

func (k *Kubernetes) createConfigMap(ctx context.Context, labels map[string]string, sshcm string, namespace string, generator *util.SSHGenerator) (configMap *coreV1.ConfigMap, err error) {

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

// ServiceHosts get service dns map
func (k *Kubernetes) ServiceHosts(ctx context.Context, namespace string) (hosts map[string]string) {
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

func waitPodReadyUsingInformer(namespace, name string, clientset kubernetes.Interface) (pod *coreV1.Pod, err error) {
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	podListener, err := clusterWatcher.PodListenerWithNamespace(clientset, namespace, stopSignal)
	if err != nil {
		return
	}
	pod = &coreV1.Pod{}
	podLabels := k8sLabels.NewSelector()
	labelKeys := []string{
		common.KTName,
	}
	requirement, err := k8sLabels.NewRequirement(labelKeys[0], selection.Equals, []string{name})
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
			p := getTargetPod(name, labelKeys, pods)
			if p != nil {
				if p.Status.Phase == "Running" {
					pod = p
					log.Info().Msgf("Shadow pod %s is ready", pod.Name)
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

// GetDeployment ...
func (k *Kubernetes) GetDeployment(ctx context.Context, name string, namespace string) (*appv1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GoPod ...
func (k *Kubernetes) GetPod(ctx context.Context, name string, namespace string) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdatePod ...
func (k *Kubernetes) UpdatePod(ctx context.Context, namespace string, pod *coreV1.Pod) (*coreV1.Pod, error) {
	return k.Clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
}

// DecreaseRef ...
func (k *Kubernetes) DecreaseRef(ctx context.Context, namespace string, app string) (cleanup bool, err error) {
	pod, err := k.GetPod(ctx, app, namespace)
	if err != nil {
		return
	}
	cleanup, err = decreaseOrRemove(ctx, k, pod)
	return
}

func decreaseOrRemove(ctx context.Context, k *Kubernetes, pod *coreV1.Pod) (cleanup bool, err error) {
	refCount := pod.ObjectMeta.Annotations[common.KTRefCount]
	if refCount == "1" {
		cleanup = true
		log.Info().Msgf("Shared shadow has only one ref, delete it")
		err = k.RemovePod(ctx, pod.GetObjectMeta().GetName(), pod.GetObjectMeta().GetNamespace())
		if err != nil {
			return
		}
	} else {
		err2 := decreasePodRef(ctx, refCount, k, pod)
		if err2 != nil {
			err = err2
			return
		}
	}
	return
}

func decreasePodRef(ctx context.Context, refCount string, k *Kubernetes, pod *coreV1.Pod) (err error) {
	log.Info().Msgf("Shared shadow has more than one ref, decrease the ref")
	count, err := decreaseRef(refCount)
	if err != nil {
		return
	}
	pod.ObjectMeta.Annotations[common.KTRefCount] = count
	_, err = k.UpdatePod(ctx, pod.GetObjectMeta().GetNamespace(), pod)
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
