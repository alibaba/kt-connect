package cluster

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func getKubernetesClient(kubeConfig string) (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	return
}

func getPodCidrs(ctx context.Context, k kubernetes.Interface, namespace, podCIDRs string) ([]string, error) {
	var cidrs []string

	if podCIDRs != "" {
		for _, cidr := range strings.Split(podCIDRs, ",") {
			cidrs = append(cidrs, cidr)
		}
		return cidrs, nil
	}

	if nodeList, err := k.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err != nil {
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
		log.Info().Msgf("Node has empty PodCIDR, try to get CIDR with pod sample")
		ipRanges, err := getPodCidrByInstance(ctx, k, namespace)
		if err != nil {
			return nil, err
		}
		for _, ir := range ipRanges {
			cidrs = append(cidrs, ir)
		}
	}

	return cidrs, nil
}

func getPodCidrByInstance(ctx context.Context, k kubernetes.Interface, namespace string) ([]string, error) {
	podList, err := k.CoreV1().Pods("").List(ctx, metav1.ListOptions{Limit: 1000})
	if err != nil {
		podList, err = k.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{Limit: 1000})
	}

	var ips []string
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return calculateMinimalIpRange(ips), nil
}

func getServiceCidr(ctx context.Context, k kubernetes.Interface, namespace string) ([]string, error) {
	serviceList, err := fetchServiceList(ctx, k, namespace)
	if err != nil {
		return []string{}, err
	}

	var ips []string
	for _, service := range serviceList.Items {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			ips = append(ips, service.Spec.ClusterIP)
		}
	}

	return calculateMinimalIpRange(ips), nil
}

// fetchServiceList try list service at cluster scope. fallback to namespace scope
func fetchServiceList(ctx context.Context, k kubernetes.Interface, namespace string) (*coreV1.ServiceList, error) {
	serviceList, err := k.CoreV1().Services("").List(ctx, metav1.ListOptions{Limit: 1000})
	if err != nil {
		return k.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{Limit: 1000})
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

func getTargetPod(labelsKey string, name string, podList []*coreV1.Pod) *coreV1.Pod {
	for _, p := range podList {
		if len(p.Labels) <= 0 {
			// almost impossible
			continue
		}
		item, containKey := p.Labels[labelsKey]
		if !containKey || item != name {
			continue
		}
		return p
	}
	return nil
}

func wait(podName string) {
	time.Sleep(3 * time.Second)
	if len(podName) > 0 {
		log.Info().Msgf("Waiting for pod %s ...", podName)
	} else {
		log.Info().Msg("Waiting for pod ...")
	}
}

func createService(metaAndSpec *SvcMetaAndSpec) *coreV1.Service {
	var servicePorts []coreV1.ServicePort
	util.MapPut(metaAndSpec.Meta.Annotations, common.KtLastHeartBeat, util.GetTimestamp())
	util.MapPut(metaAndSpec.Meta.Labels, common.ControlBy, common.KubernetesTool)

	for srcPort, targetPort := range metaAndSpec.Ports {
		servicePorts = append(servicePorts, coreV1.ServicePort{
			Name:       fmt.Sprintf("%s-%d", metaAndSpec.Meta.Name, srcPort),
			Port:       int32(srcPort),
			TargetPort: intstr.FromInt(targetPort),
		})
	}

	service := &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        metaAndSpec.Meta.Name,
			Namespace:   metaAndSpec.Meta.Namespace,
			Labels:      metaAndSpec.Meta.Labels,
			Annotations: metaAndSpec.Meta.Annotations,
		},
		Spec: coreV1.ServiceSpec{
			Selector: metaAndSpec.Selectors,
			Type:     coreV1.ServiceTypeClusterIP,
			Ports:    servicePorts,
		},
	}
	if metaAndSpec.External {
		service.Spec.Type = coreV1.ServiceTypeLoadBalancer
	}
	return service
}

func createContainer(image string, args []string, envs map[string]string, options *options.DaemonOptions) coreV1.Container {
	var envVar []coreV1.EnvVar
	for k, v := range envs {
		envVar = append(envVar, coreV1.EnvVar{Name: k, Value: v})
	}
	var pullPolicy coreV1.PullPolicy
	if options.AlwaysUpdateShadow {
		pullPolicy = "Always"
	} else {
		pullPolicy = "IfNotPresent"
	}
	return coreV1.Container{
		Name:            common.DefaultContainer,
		Image:           image,
		ImagePullPolicy: pullPolicy,
		Args:            args,
		Env:             envVar,
		SecurityContext: &coreV1.SecurityContext{
			Capabilities: &coreV1.Capabilities{
				Add: []coreV1.Capability{
					"AUDIT_WRITE",
				},
			},
		},
	}
}

func createPod(metaAndSpec *PodMetaAndSpec, options *options.DaemonOptions) *coreV1.Pod {
	var args []string
	namespace := metaAndSpec.Meta.Namespace
	name := metaAndSpec.Meta.Name
	labels := metaAndSpec.Meta.Labels
	annotations := metaAndSpec.Meta.Annotations
	annotations[common.KtRefCount] = "1"
	annotations[common.KtLastHeartBeat] = util.GetTimestamp()
	image := metaAndSpec.Image
	envs := metaAndSpec.Envs

	pod := &coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: coreV1.PodSpec{
			ServiceAccountName: options.ServiceAccount,
			Containers: []coreV1.Container{
				createContainer(image, args, envs, options),
			},
		},
	}

	if options.ImagePullSecret != "" {
		addImagePullSecret(pod, options.ImagePullSecret)
	}

	if options.NodeSelector != "" {
		pod.Spec.NodeSelector = util.String2Map(options.NodeSelector)
	}

	return pod
}

func getSSHVolume(volume string) coreV1.Volume {
	sshVolume := coreV1.Volume{
		Name: "ssh-public-key",
		VolumeSource: coreV1.VolumeSource{
			ConfigMap: &coreV1.ConfigMapVolumeSource{
				LocalObjectReference: coreV1.LocalObjectReference{
					Name: volume,
				},
				Items: []coreV1.KeyToPath{
					{
						Key:  common.SshAuthKey,
						Path: "authorized_keys",
					},
				},
			},
		},
	}
	return sshVolume
}

func addImagePullSecret(pod *coreV1.Pod, imagePullSecret string) {
	pod.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{
		{
			Name: imagePullSecret,
		},
	}
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}

func decreaseRef(refCount string) (count string, err error) {
	currentCount, err := strconv.Atoi(refCount)
	if err != nil {
		return
	}
	count = strconv.Itoa(currentCount - 1)
	return
}
