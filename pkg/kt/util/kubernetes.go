package util

import (
	"fmt"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetCirds Get kubernetes cluster resource crids
func GetCirds(clientset *kubernetes.Clientset, podCIDR string) (cidrs []string, err error) {
	cidrs, err = getPodCirds(clientset, podCIDR)
	if err != nil {
		return
	}
	serviceCird, err := getServiceCird(clientset)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCird...)
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
		log.Info().Msgf("Fail to get pod cidr from node.Spec.PODCIDR, try to get with pod sample")
		podList, err2 := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err2 != nil {
			log.Printf("Fails to get service info of cluster")
			return
		}

		samples := mapset.NewSet()
		for _, pod := range podList.Items {
			if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
				samples.Add(getCirdFromSample(pod.Status.PodIP))
			}
		}

		for _, sample := range samples.ToSlice() {
			cidrs = append(cidrs, fmt.Sprint(sample))
		}
	}

	return
}

func getServiceCird(clientset *kubernetes.Clientset) (cidr []string, err error) {
	serviceList, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		log.Printf("Fails to get service info of cluster")
		return cidr, err
	}

	samples := mapset.NewSet()
	for _, service := range serviceList.Items {
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
