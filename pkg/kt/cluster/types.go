package cluster

import (
	"fmt"
	"strings"

	clusterWatcher "github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	mapset "github.com/deckarep/golang-set"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
)

// KubernetesFactory kubernetes factory
type KubernetesFactory struct {
}

// Create kubernetes instance
func (f *KubernetesFactory) Create(kubeConfig string) (kubernetes Kubernetes, err error) {
	clientSet, err := GetKubernetesClient(kubeConfig)
	if err != nil {
		return
	}
	serviceListener, err := clusterWatcher.ServiceListener(clientSet)
	podListener, err := clusterWatcher.PodListener(clientSet)
	if err != nil {
		return
	}
	kubernetes = Kubernetes{
		Clientset:       clientSet,
		ServiceListener: serviceListener,
		PodListener:     podListener,
	}
	return
}

// KubernetesInterface kubernetes interface
type KubernetesInterface interface {
	Deployment(name, namespace string) (deployment appV1.Deployment, err error)
	Scale(name, namespace string, replicas *int32) (err error)
	ServiceHosts(namespace string) (hosts map[string]string)
	ClusterCrids(podCIDR string) (cidrs []string, err error)
	CreateShadow(name, namespace, image string, labels map[string]string) (podIP, podName string, err error)
}

// Kubernetes implements KubernetesInterface
type Kubernetes struct {
	Clientset       *kubernetes.Clientset
	ServiceListener v1.ServiceLister
	PodListener     v1.PodLister
}

// Scale scale deployment to
func (k *Kubernetes) Scale(deployment *appV1.Deployment, replicas *int32) (err error) {
	log.Printf("scale deployment %s to %d\n", deployment.GetObjectMeta().GetName(), *replicas)
	client := k.Clientset.AppsV1().Deployments(deployment.GetObjectMeta().GetNamespace())
	deployment.Spec.Replicas = replicas

	d, err := client.Update(deployment)
	if err != nil {
		log.Printf("%s Fails scale deployment %s to %d\n", err.Error(), deployment.GetObjectMeta().GetName(), *replicas)
		return
	}
	log.Printf(" * %s (%d replicas) success", d.Name, *d.Spec.Replicas)
	return
}

// Deployment get deployment
func (k *Kubernetes) Deployment(name, namespace string) (deployment *appV1.Deployment, err error) {
	deployment, err = k.Clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	return
}

// CreateShadow create shadow
func (k *Kubernetes) CreateShadow(name, namespace, image string, labels map[string]string) (podIP, podName string, err error) {
	return CreateShadow(k.Clientset, name, labels, namespace, image)
}

// ClusterCrids get cluster cirds
func (k *Kubernetes) ClusterCrids(podCIDR string) (cidrs []string, err error) {
	serviceList, err := k.ServiceListener.List(labels.Everything())
	if err != nil {
		return
	}

	cidrs, err = getPodCirds(k.Clientset, podCIDR)
	if err != nil {
		return
	}

	serviceCird, err := getServiceCird(serviceList)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCird...)
	return
}

// ServiceHosts get service dns map
func (k *Kubernetes) ServiceHosts(namespace string) (hosts map[string]string) {
	services, err := k.ServiceListener.Services(namespace).List(labels.Everything())
	if err != nil {
		return
	}
	hosts = map[string]string{}
	for _, service := range services {
		hosts[service.ObjectMeta.Name] = service.Spec.ClusterIP
	}
	return
}

func getPodCirds(clientset *kubernetes.Clientset, podCIDR string) (cidrs []string, err error) {
	cidrs = []string{}

	if len(podCIDR) != 0 {
		cidrs = append(cidrs, podCIDR)
		return
	}

	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})

	if err != nil {
		log.Printf("Fails to get node info of cluster")
		return nil, err
	}

	for _, node := range nodeList.Items {
		if node.Spec.PodCIDR != "" && len(node.Spec.PodCIDR) != 0 {
			cidrs = append(cidrs, node.Spec.PodCIDR)
		}
	}

	if len(cidrs) == 0 {
		samples, err2 := getPodCirdByInstance(clientset)
		if err2 != nil {
			err = err2
			return
		}
		for _, sample := range samples.ToSlice() {
			cidrs = append(cidrs, fmt.Sprint(sample))
		}
	}

	return
}

func getPodCirdByInstance(clientset *kubernetes.Clientset) (samples mapset.Set, err error) {
	log.Info().Msgf("Fail to get pod cidr from node.Spec.PODCIDR, try to get with pod sample")
	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Fails to get service info of cluster")
		return
	}

	samples = mapset.NewSet()
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
			samples.Add(getCirdFromSample(pod.Status.PodIP))
		}
	}
	return
}

func getServiceCird(serviceList []*coreV1.Service) (cidr []string, err error) {
	samples := mapset.NewSet()
	for _, service := range serviceList {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			samples.Add(getCirdFromSample(service.Spec.ClusterIP))
		}
	}

	for _, sample := range samples.ToSlice() {
		cidr = append(cidr, fmt.Sprint(sample))
	}
	return
}

func getCirdFromSample(sample string) string {
	return strings.Join(append(strings.Split(sample, ".")[:2], []string{"0", "0"}...), ".") + "/16"
}
