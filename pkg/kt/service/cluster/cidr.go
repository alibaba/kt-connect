package cluster

import (
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(namespace string) ([]string, error) {
	ips := getServiceIps(k.Clientset, namespace)
	if !opt.Get().ConnectOptions.DisablePodIp {
		ips = append(ips, getPodIps(k.Clientset, namespace)...)
	}

	cidrs := calculateMinimalIpRange(ips)
	log.Debug().Msgf("Cluster CIDR are: %v", cidrs)

	apiServerIp := util.ExtractHostIp(opt.Get().RuntimeStore.RestConfig.Host)
	log.Debug().Msgf("Using cluster ip %s", apiServerIp)
	if apiServerIp != "" {
		cidrs = removeCidrOf(cidrs, apiServerIp + "/32")
	}

	if opt.Get().ConnectOptions.IncludeIps != "" {
		for _, ipRange := range strings.Split(opt.Get().ConnectOptions.IncludeIps, ",") {
			if opt.Get().ConnectOptions.Mode == util.ConnectModeTun2Socks && isSingleIp(ipRange) {
				log.Warn().Msgf("Includes single IP '%s' is not allow in %s mode", ipRange, util.ConnectModeTun2Socks)
			} else {
				cidrs = append(cidrs, ipRange)
			}
		}
	}
	if opt.Get().ConnectOptions.ExcludeIps != "" {
		for _, ipRange := range strings.Split(opt.Get().ConnectOptions.ExcludeIps, ",") {
			cidrs = util.ArrayDelete(cidrs, ipRange)
		}
	}
	return cidrs, nil
}

func removeCidrOf(cidrRanges []string, ipRange string) []string {
	var newRange []string
	for _, cidr := range cidrRanges {
		if !isPartOfRange(cidr, ipRange) {
			newRange = append(newRange, cidr)
		}
	}
	return newRange
}

func isPartOfRange(ipRange string, subIpRange string) bool {
	ipRangeBin, err := ipRangeToBin(ipRange)
	if err != nil {
		return false
	}
	subIpRangeBin, err := ipRangeToBin(subIpRange)
	if err != nil {
		return false
	}
	for i := 0; i < 32; i++ {
		if ipRangeBin[i] == -1 {
			return true
		}
		if subIpRangeBin[i] != ipRangeBin[i] {
			return false
		}
	}
	return true
}

func getPodIps(k kubernetes.Interface, namespace string) []string {
	podList, err := k.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{Limit: 1000})
	if err != nil {
		podList, err = k.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1000})
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to fetch pod ips")
			return []string{}
		}
	}

	var ips []string
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return ips
}

func getServiceIps(k kubernetes.Interface, namespace string) []string {
	serviceList, err := k.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{Limit: 1000})
	if err != nil {
		serviceList, err = k.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{Limit: 1000})
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to fetch service ips")
			return []string{}
		}
	}

	var ips []string
	for _, service := range serviceList.Items {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			ips = append(ips, service.Spec.ClusterIP)
		}
	}

	return ips
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
						// partially equal and over threshold, mark the match start position
						match = true
						miniBins[i][j] = -1
					}
					break
				} else if j == 31 {
					// fully equal
					match = true
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

func ipRangeToBin(ipRange string) ([32]int, error) {
	parts := strings.Split(ipRange, "/")
	if len(parts) != 2 {
		return [32]int{}, fmt.Errorf("invalid ip range format: %s", ipRange)
	}
	sepIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return [32]int{}, err
	}
	ipBin, err := ipToBin(parts[0])
	if err != nil {
		return [32]int{}, err
	}
	if sepIndex < 32 {
		ipBin[sepIndex] = -1
	}
	return ipBin, nil
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
