package cluster

import (
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(namespace string) (cidrs []string, err error) {
	if !opt.Get().ConnectOptions.DisablePodIp {
		cidrs, err = getPodCidrs(k.Clientset, namespace)
		if err != nil {
			return
		}
	}
	log.Debug().Msgf("Pod CIDR is %v", cidrs)

	serviceCidr, err := getServiceCidr(k.Clientset, namespace)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCidr...)
	log.Debug().Msgf("Service CIDR is %v", serviceCidr)

	if opt.Get().ConnectOptions.IncludeIps != "" {
		for _, ipRange := range strings.Split(opt.Get().ConnectOptions.IncludeIps, ",") {
			if opt.Get().ConnectOptions.Mode == util.ConnectModeTun2Socks && isSingleIp(ipRange) {
				log.Warn().Msgf("Includes single IP '%s' is not allow in %s mode", ipRange, util.ConnectModeTun2Socks)
			} else {
				cidrs = append(cidrs, ipRange)
			}
		}
	}
	return calculateMinimalIpRange(cidrs), nil
}

func getPodCidrs(k kubernetes.Interface, namespace string) ([]string, error) {
	var cidrs []string
	if nodeList, err := k.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{}); err != nil {
		// Usually cause by the local kube config has not enough permission
		log.Debug().Err(err).Msgf("Failed to read node information of cluster")
	} else {
		for _, node := range nodeList.Items {
			if node.Spec.PodCIDR != "" && len(node.Spec.PodCIDR) != 0 {
				cidrs = append(cidrs, node.Spec.PodCIDR)
			}
		}
	}

	if len(cidrs) == 0 {
		log.Info().Msgf("No PodCIDR available, fetching with pod samples")
		ipRanges, err := getPodCidrByInstance(k, namespace)
		if err != nil {
			return nil, err
		}
		for _, ir := range ipRanges {
			cidrs = append(cidrs, ir)
		}
	}

	return cidrs, nil
}

func getPodCidrByInstance(k kubernetes.Interface, namespace string) ([]string, error) {
	podList, err := k.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{Limit: 1000})
	if err != nil {
		podList, err = k.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1000})
	}

	var ips []string
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return ips, nil
}

func getServiceCidr(k kubernetes.Interface, namespace string) ([]string, error) {
	serviceList, err := fetchServiceList(k, namespace)
	if err != nil {
		return []string{}, err
	}

	var ips []string
	for _, service := range serviceList.Items {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			ips = append(ips, service.Spec.ClusterIP)
		}
	}

	return ips, nil
}

// fetchServiceList try list service at cluster scope. fallback to namespace scope
func fetchServiceList(k kubernetes.Interface, namespace string) (*coreV1.ServiceList, error) {
	serviceList, err := k.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{Limit: 1000})
	if err != nil {
		return k.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1000})
	}
	return serviceList, err
}

func calculateMinimalIpRange(ips []string) []string {
	var miniBins [][32]int
	threshold := 16
	withAlign := true
	for _, ip := range ips {
		ipBin, err := ipToBin(ip)
		if err != nil {
			// skip invalid ip
			continue
		}
		if len(miniBins) == 0 {
			// accept first ip
			miniBins = append(miniBins, ipBin)
			continue
		}
		match := false
		for i, bins := range miniBins {
			for j, b := range bins {
				if b != ipBin[j] {
					if j >= threshold {
						// mark the match start position
						match = true
						miniBins[i][j] = -1
					}
					break
				}
			}
			if match {
				break
			}
		}
		if !match {
			// no include in current range, append it
			miniBins = append(miniBins, ipBin)
		}
	}
	var miniRange []string
	for _, bins := range miniBins {
		miniRange = append(miniRange, binToIpRange(bins, withAlign))
	}
	return miniRange
}

func binToIpRange(bins [32]int, withAlign bool) string {
	ips := []string {"0", "0", "0", "0"}
	mask := 0
	end := false
	for i := 0; i < 4; i++ {
		segment := 0
		factor := 128
		for j := 0; j < 8; j++ {
			if bins[i*8+j] < 0 {
				end = true
				break
			}
			segment += bins[i*8+j] * factor
			factor /= 2
			mask++
		}
		if !withAlign || !end {
			ips[i] = strconv.Itoa(segment)
		}
		if end {
			if withAlign {
				mask = i * 8
			}
			break
		}
	}
	return fmt.Sprintf("%s/%d", strings.Join(ips, "."), mask)
}

func ipToBin(ip string) (ipBin [32]int, err error) {
	ipNum, err := parseIp(ip)
	if err != nil {
		return
	}
	for i, n := range ipNum {
		bin := decToBin(n)
		copy(ipBin[i*8:i*8+8], bin[:])
	}
	return
}

func parseIp(ip string) (ipNum [4]int, err error) {
	for i, seg := range strings.Split(ip, ".") {
		ipNum[i], err = strconv.Atoi(seg)
		if err != nil {
			return
		}
	}
	return
}

func decToBin(n int) [8]int {
	var bin [8]int
	for i := 0; n > 0; n /= 2 {
		bin[i] = n % 2
		i++
	}
	// revert it
	for i, j := 0, len(bin)-1; i < j; i, j = i+1, j-1 {
		bin[i], bin[j] = bin[j], bin[i]
	}
	return bin
}
