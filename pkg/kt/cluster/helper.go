package cluster

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient get Kubernetes client from config
func GetKubernetesClient(kubeConfig string) (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
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
		log.Error().Msgf("Fails to get node info of cluster")
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
		log.Error().Msg("Fails to get service info of cluster")
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
