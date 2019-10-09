package util

import (
	"fmt"
	"math/rand"
	"net"
	"strings"

	"github.com/deckarep/golang-set"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (address string) {
	address = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}

// RandomString Generate RandomString
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

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
			sample := strings.Join(append(strings.Split(service.Spec.ClusterIP, ".")[:2], []string{"0", "0"}...), ".") + "/16"
			samples.Add(sample)
		}
	}

	for _, sample := range samples.ToSlice() {
		cidr = append(cidr, fmt.Sprint(sample))
	}

	return
}
