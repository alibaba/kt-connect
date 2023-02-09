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

// ClusterCidr get cluster CIDR
func (k *Kubernetes) ClusterCidr(namespace string) ([]string, []string) {
	ips := getServiceIps(k.Clientset, namespace)
	log.Debug().Msgf("Found %d IPs", len(ips))
	svcCidr := calculateMinimalIpRange(ips)
	log.Debug().Msgf("Service ips are: %v", ips)
	log.Debug().Msgf("Service CIDR are: %v", svcCidr)

	var podCidr []string
	if !opt.Get().Connect.DisablePodIp {
		ips = getPodIps(k.Clientset, namespace)
		log.Debug().Msgf("Found %d IPs", len(ips))
		podCidr = calculateMinimalIpRange(ips)
		log.Debug().Msgf("Pod ips are: %v", ips)
		log.Debug().Msgf("Pod CIDR are: %v", podCidr)
	}

	apiServerIp := util.ExtractHostIp(opt.Store.RestConfig.Host)
	log.Debug().Msgf("Using cluster IP %s", apiServerIp)

	if opt.Store.Ipv6Cluster == true  && strings.Contains(apiServerIp, ":"){
		apiServerIp = strings.Split(strings.Split(opt.Store.RestConfig.Host, "[")[1], "]")[0]
	}

	cidr := mergeIpRange(svcCidr, podCidr, apiServerIp)
	log.Debug().Msgf("Cluster CIDR are: %v", cidr)

	excludeIps := strings.Split(opt.Get().Connect.ExcludeIps, ",")
	var excludeCidr []string
	if len(apiServerIp) > 0 {
		excludeIps = append(excludeIps, apiServerIp+"/32")
		log.Debug().Msgf("excludeIps are: %v", excludeIps)
	}

	if opt.Get().Connect.IncludeIps != "" {
		for _, ipRange := range strings.Split(opt.Get().Connect.IncludeIps, ",") {
			if opt.Get().Connect.Mode == util.ConnectModeTun2Socks && isSingleIp(ipRange) {
				log.Warn().Msgf("Includes single IP '%s' is not allow in %s mode", ipRange, util.ConnectModeTun2Socks)
			} else {
				cidr = append(cidr, ipRange)
			}
		}
	}
	if opt.Get().Connect.ExcludeIps != "" {
		for _, ipRange := range excludeIps {
			var toRemove []string
			for _, r := range cidr {
				if r == ipRange {
					// if exclude ip equal to cidr, remove it and break
					toRemove = append(toRemove, r)
					break
				} else if isPartOfRange(ipRange, r) {
					// if exclude ip overlap cidr, remove it
					toRemove = append(toRemove, r)
				} else if isPartOfRange(r, ipRange) {
					// if cidr overlap exclude ip, should set bypass route for it
					excludeCidr = append(excludeCidr, ipRange)
					break
				}
				// otherwise, exclude ip not part of cidr, ignore it
			}
			for _, r := range toRemove {
				cidr = util.ArrayDelete(cidr, r)
			}
		}
	}
	if len(excludeCidr) > 0 {
		log.Debug().Msgf("Non-cluster CIDR are: %v", excludeCidr)
	}

	// remove ipv6 api address
	if opt.Store.Ipv6Cluster == true  && strings.Contains(apiServerIp, ":") {
		s := strings.Split(apiServerIp, ":")
		ipmask := fmt.Sprintf("%s:%s::/32", s[0], s[1])
		cidr = util.ArrayDelete(cidr, ipmask)
	}
	return cidr, excludeCidr
}

func mergeIpRange(svcCidr []string, podCidr []string, apiServerIp string) []string {
	cidr := calculateMinimalIpRange(append(svcCidr, podCidr...))
	mergedCidr := make([]string, 0)
	for _, r := range cidr {
		if isPartOfRange(r, apiServerIp+"/32") {
			mergedCidr = append(mergedCidr, excludeIpFromRange(r, apiServerIp+"/32")...)
		} else {
			mergedCidr = append(mergedCidr, r)
		}
	}
	return mergedCidr
}

func excludeIpFromRange(ipRange string, ip string) []string {
	ipRangeBin, err := ipRangeToBin(ipRange)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to parse IP range '%s', skipping", ipRange)
		return []string{}
	}
	ipBin, err := ipToBin(ip)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to parse IP '%s', skipping", ip)
		return []string{}
	}
	var excludedRanges []string
	var excludedIpBin [32]int
	reachHostPart := false
	for i := 0; i < 32; i++ {
		if i < 31 && ipRangeBin[i+1] == -1 {
			reachHostPart = true
		} else if reachHostPart {
			excludedIpBin[i-1] = ipBin[i-1]
			if ipBin[i] == 0 {
				excludedIpBin[i] = 1
			} else {
				excludedIpBin[i] = 0
			}
			if i < 31 {
				excludedIpBin[i+1] = -1
			}
			excludedRanges = append(excludedRanges, binToIpRange(excludedIpBin, false))
		} else if ipRangeBin[i] != ipBin[i] {
			log.Warn().Err(err).Msgf("IP '%s' is not part of range '%s', skipping", ipRange, ip)
			return []string{}
		} else {
			excludedIpBin[i] = ipBin[i]
		}
	}
	return excludedRanges
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
	podList, err := k.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		Limit:          1000,
		TimeoutSeconds: &apiTimeout,
	})
	if err != nil {
		podList, err = k.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			Limit:          1000,
			TimeoutSeconds: &apiTimeout,
		})
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
	serviceList, err := k.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{
		Limit:          1000,
		TimeoutSeconds: &apiTimeout,
	})
	if err != nil {
		serviceList, err = k.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
			Limit:          1000,
			TimeoutSeconds: &apiTimeout,
		})
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

func calculateMinimalIpv6Range(ips []string) []string {
	var miniRange []string
	for _, ip := range ips {
		if strings.Contains(ip, ".") {
			continue
		}
		s := strings.Split(ip, ":")
		ipmask := fmt.Sprintf("%s:%s::/32", s[0], s[1])
		if !util.Contains(miniRange, ipmask) {
			miniRange = append(miniRange, ipmask)
		}

	}
	return miniRange
}

func calculateMinimalIpRange(ips []string) []string {
	if opt.Store.Ipv6Cluster == true {
		return calculateMinimalIpv6Range(ips)
	}

	var miniBins [][32]int
	threshold := 16
	withAlign := true
	for _, ip := range ips {
		if strings.Contains(ip, ":") {
			continue
		}
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
	ips := []string{"0", "0", "0", "0"}
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
	slashCount := strings.Count(ip, "/")
	if slashCount == 1 {
		return ipRangeToBin(ip)
	} else if slashCount > 1 {
		err = fmt.Errorf("invalid ip address: %s", ip)
		return
	}
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
