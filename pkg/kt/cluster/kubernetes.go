package cluster

import (
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"strconv"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/options"

	clusterWatcher "github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"k8s.io/client-go/kubernetes"
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
	Sshcm          string
	PrivateKeyPath string
}

// RemoveService remove service
func (k *Kubernetes) RemoveService(name, namespace string) (err error) {
	client := k.Clientset.CoreV1().Services(namespace)
	return client.Delete(name, &metav1.DeleteOptions{})
}

// RemoveDeployment remove deployment instances
func (k *Kubernetes) RemoveDeployment(name, namespace string) (err error) {
	deploymentsClient := k.Clientset.AppsV1().Deployments(namespace)
	deletePolicy := metav1.DeletePropagationBackground
	return deploymentsClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// RemoveConfigMap remove ConfigMap instance
func (k *Kubernetes) RemoveConfigMap(name, namespace string) (err error) {
	cli := k.Clientset.CoreV1().ConfigMaps(namespace)
	deletePolicy := metav1.DeletePropagationBackground
	return cli.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// ScaleTo scale deployment to
func (k *Kubernetes) ScaleTo(deployment, namespace string, replicas *int32) (err error) {
	obj, err := k.Deployment(deployment, namespace)
	if err != nil {
		return
	}
	return k.Scale(obj, replicas)
}

// Scale scale deployment to
func (k *Kubernetes) Scale(deployment *appv1.Deployment, replicas *int32) (err error) {
	log.Info().Msgf("Scaling deployment %s to %d", deployment.GetObjectMeta().GetName(), *replicas)
	client := k.Clientset.AppsV1().Deployments(deployment.GetObjectMeta().GetNamespace())
	deployment.Spec.Replicas = replicas

	d, err := client.Update(deployment)
	if err != nil {
		log.Error().Msgf("Fails scale deployment %s to %d: %s", deployment.GetObjectMeta().GetName(), *replicas, err.Error())
		return
	}
	log.Info().Msgf(" * %s (%d replicas) success", d.Name, *d.Spec.Replicas)
	return
}

// Deployment get deployment
func (k *Kubernetes) Deployment(name, namespace string) (*appv1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

// GetOrCreateShadow create shadow
func (k *Kubernetes) GetOrCreateShadow(name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (
	podIP, podName, sshcm string, credential *util.SSHCredential, err error) {

	component := labels[common.KTComponent]
	identifier := strings.ToLower(util.RandomString(4))
	sshcm = fmt.Sprintf("kt-%s-public-key-%s", component, identifier)

	privateKeyPath := util.PrivateKeyPath(component, identifier)

	if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
		pod, generator, err2 := k.tryGetExistingShadowRelatedObjs(&ResourceMeta{
			Name:        name,
			Namespace:   options.Namespace,
			Labels:      labels,
			Annotations: annotations,
		}, &SSHkeyMeta{
			Sshcm:          sshcm,
			PrivateKeyPath: privateKeyPath,
		})
		if err2 != nil {
			err = err2
			return
		}
		if pod != nil && generator != nil {
			podIP, podName, credential = shadowResult(*pod, generator)
			return
		}
	}

	podIP, podName, credential, err = k.createShadow(&PodMetaAndSpec{
		&ResourceMeta{
			Name:        name,
			Namespace:   options.Namespace,
			Labels:      labels,
			Annotations: annotations,
		}, options.Image, envs,
	}, &SSHkeyMeta{
		Sshcm:          sshcm,
		PrivateKeyPath: privateKeyPath,
	}, options)
	return
}

func (k *Kubernetes) createShadow(metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta, options *options.DaemonOptions) (
	podIP string, podName string, credential *util.SSHCredential, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}
	configMap, err2 := k.createConfigMap(metaAndSpec.Meta.Labels, sshKeyMeta.Sshcm, metaAndSpec.Meta.Namespace, generator)

	if err2 != nil {
		err = err2
		return
	}
	log.Info().Msgf("Successful create ssh config map %v", configMap.ObjectMeta.Name)

	pod, err2 := k.createAndGetPod(metaAndSpec, sshKeyMeta.Sshcm, options)
	if err2 != nil {
		err = err2
		return
	}
	podIP, podName, credential = shadowResult(pod, generator)
	return
}

// GetAllExistingShadowDeployments fetch all shadow deployments
func (k *Kubernetes) GetAllExistingShadowDeployments(namespace string) ([]appv1.Deployment, error) {
	list, err := k.Clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: k8sLabels.Set(metav1.LabelSelector{
			MatchLabels: map[string]string{common.ControlBy: common.KubernetesTool},
		}.MatchLabels).String(),
	})
	if list == nil {
		return nil, common.CommandExecError{Reason: "get nil list when querying shadow deployments"}
	}
	return list.Items, err
}

func (k *Kubernetes) tryGetExistingShadowRelatedObjs(resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (pod *v1.Pod, generator *util.SSHGenerator, err error) {
	_, shadowError := k.GetDeployment(resourceMeta.Name, resourceMeta.Namespace)
	if shadowError != nil {
		return
	}
	cli := k.Clientset.CoreV1().ConfigMaps(resourceMeta.Namespace)
	configMap, configMapError := cli.Get(sshKeyMeta.Sshcm, metav1.GetOptions{})

	if configMapError != nil {
		err = errors.New("Found shadow deployment but no configMap. Please delete the deployment " + resourceMeta.Name)
		return
	}

	generator = util.NewSSHGenerator(configMap.Data[common.SSHAuthPrivateKey], configMap.Data[common.SSHAuthKey], sshKeyMeta.PrivateKeyPath)

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SSHAuthPrivateKey]))
	if err != nil {
		return
	}

	return k.getShadowPod(resourceMeta, generator)
}

func (k *Kubernetes) getShadowPod(resourceMeta *ResourceMeta, generator *util.SSHGenerator) (pod *v1.Pod, sshGenerator *util.SSHGenerator, err error) {
	podList, err := k.Clientset.CoreV1().Pods(resourceMeta.Namespace).List(metav1.ListOptions{
		LabelSelector: k8sLabels.Set(metav1.LabelSelector{MatchLabels: resourceMeta.Labels}.MatchLabels).String(),
	})
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shared shadow, reuse it")
		err = increaseRefCount(resourceMeta.Name, k.Clientset, resourceMeta.Namespace)
		if err != nil {
			return
		}
		return &(podList.Items[0]), generator, nil
	}
	if len(podList.Items) > 1 {
		err = errors.New("Found more than one pod with name " + resourceMeta.Name + ", please make sure these is only one in namespace " + resourceMeta.Namespace)
	} else {
		err = errors.New("no Shadow pod found while shadow deployment present. You may need to clean up the deployment by yourself")
	}
	return
}

func increaseRefCount(name string, clientSet kubernetes.Interface, namespace string) error {
	deployment, err := clientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := deployment.ObjectMeta.Annotations
	count, err := strconv.Atoi(annotations[common.KTRefCount])
	if err != nil {
		log.Error().Msgf("Failed to parse annotations[common.KTRefCount] of deployment %s with value %s", name, annotations[common.KTRefCount])
		return err
	}

	deployment.ObjectMeta.Annotations[common.KTRefCount] = strconv.Itoa(count + 1)

	_, err = clientSet.AppsV1().Deployments(namespace).Update(deployment)
	return err
}

func shadowResult(pod v1.Pod, generator *util.SSHGenerator) (string, string, *util.SSHCredential) {
	podIP := pod.Status.PodIP
	podName := pod.GetObjectMeta().GetName()
	credential := util.NewDefaultSSHCredential()
	credential.PrivateKeyPath = generator.PrivateKeyPath
	return podIP, podName, credential
}

func (k *Kubernetes) createAndGetPod(metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) (pod v1.Pod, err error) {
	localIPAddress := util.GetOutboundIP()
	log.Info().Msgf("Client address %s", localIPAddress)
	resourceMeta := metaAndSpec.Meta
	resourceMeta.Labels[common.KTRemoteAddress] = localIPAddress
	resourceMeta.Labels[common.KTName] = resourceMeta.Name
	cli := k.Clientset.AppsV1().Deployments(resourceMeta.Namespace)
	util.SetupDeploymentHeartBeat(cli, resourceMeta.Name)

	deployment := deployment(metaAndSpec, sshcm, options)
	log.Info().Msg("Shadow template is prepare ready.")
	result, err := cli.Create(deployment)
	if err != nil {
		return
	}
	log.Info().Msgf("Deploy shadow deployment %s in namespace %s", result.GetObjectMeta().GetName(), resourceMeta.Namespace)

	return waitPodReadyUsingInformer(resourceMeta.Namespace, resourceMeta.Name, k.Clientset)
}

func (k *Kubernetes) createConfigMap(labels map[string]string, sshcm string, namespace string, generator *util.SSHGenerator) (configMap *v1.ConfigMap, err error) {

	annotations := map[string]string{common.KTLastHeartBeat: util.GetTimestamp()}
	labels[common.KTName] = sshcm
	cli := k.Clientset.CoreV1().ConfigMaps(namespace)
	util.SetupConfigMapHeartBeat(cli, sshcm)

	return cli.Create(&v1.ConfigMap{
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
	})
}

// CreateService create kubernetes service
func (k *Kubernetes) CreateService(name, namespace string, external bool, port int, labels map[string]string) (*v1.Service, error) {
	cli := k.Clientset.CoreV1().Services(namespace)
	util.SetupServiceHeartBeat(cli, name)
	svc := service(name, namespace, labels, external, port)
	return cli.Create(svc)
}

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(namespace string, connectOptions *options.ConnectOptions) (cidrs []string, err error) {
	currentNS := namespace
	if connectOptions.Global {
		log.Info().Msgf("Scan proxy CIDR in cluster scope")
		currentNS = ""
	} else {
		log.Info().Msgf("Scan proxy CIDR in namespace scope")
	}

	serviceList, err := k.Clientset.CoreV1().Services(currentNS).List(metav1.ListOptions{})
	if err != nil {
		return
	}

	cidrs, err = getPodCidrs(k.Clientset, connectOptions.CIDR)
	if err != nil {
		return
	}

	services := serviceList.Items
	serviceCidr, err := getServiceCidr(services)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCidr...)
	return
}

// ServiceHosts get service dns map
func (k *Kubernetes) ServiceHosts(namespace string) (hosts map[string]string) {
	services, err := k.Clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	hosts = map[string]string{}
	for _, service := range services.Items {
		hosts[service.ObjectMeta.Name] = service.Spec.ClusterIP
	}
	return
}

func waitPodReadyUsingInformer(namespace, name string, clientset kubernetes.Interface) (pod v1.Pod, err error) {
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	podListener, err := clusterWatcher.PodListenerWithNamespace(clientset, namespace, stopSignal)
	if err != nil {
		return
	}
	pod = v1.Pod{}
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
wait_loop:
	for {
		hasRunningPod := len(pods) > 0
		var podName string
		if hasRunningPod {
			// podLister do not support FieldSelector
			// https://github.com/kubernetes/client-go/issues/604
			p := getTargetPod(name, labelKeys, pods)
			if p != nil {
				if p.Status.Phase == "Running" {
					pod = *p
					log.Info().Msgf("Shadow pod: %s is ready.", pod.Name)
					break wait_loop
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
func (k *Kubernetes) GetDeployment(name string, namespace string) (*appv1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

// UpdateDeployment ...
func (k *Kubernetes) UpdateDeployment(namespace string, deployment *appv1.Deployment) (*appv1.Deployment, error) {
	return k.Clientset.AppsV1().Deployments(namespace).Update(deployment)
}

// DecreaseRef ...
func (k *Kubernetes) DecreaseRef(namespace string, app string) (cleanup bool, err error) {
	deployment, err := k.GetDeployment(app, namespace)
	if err != nil {
		return
	}
	cleanup, err = decreaseOrRemove(k, deployment)
	return
}

func decreaseOrRemove(k *Kubernetes, deployment *appv1.Deployment) (cleanup bool, err error) {
	refCount := deployment.ObjectMeta.Annotations[common.KTRefCount]
	if refCount == "1" {
		cleanup = true
		log.Info().Msgf("Shared shadow has only one ref, delete it")
		err = k.RemoveDeployment(deployment.GetObjectMeta().GetName(), deployment.GetObjectMeta().GetNamespace())
		if err != nil {
			return
		}
	} else {
		err2 := decreaseDeploymentRef(refCount, k, deployment)
		if err2 != nil {
			err = err2
			return
		}
	}
	return
}

func decreaseDeploymentRef(refCount string, k *Kubernetes, deployment *appv1.Deployment) (err error) {
	log.Info().Msgf("Shared shadow has more than one ref, decrease the ref")
	count, err := decreaseRef(refCount)
	if err != nil {
		return
	}
	deployment.ObjectMeta.Annotations[common.KTRefCount] = count
	_, err = k.UpdateDeployment(deployment.GetObjectMeta().GetNamespace(), deployment)
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
