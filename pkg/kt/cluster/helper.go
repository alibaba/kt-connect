package cluster

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	mapset "github.com/deckarep/golang-set"
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

func getPodCidrs(ctx context.Context, clientset kubernetes.Interface, podCIDRs string) ([]string, error) {
	var cidrs []string

	if podCIDRs != "" {
		for _, cidr := range strings.Split(podCIDRs, ",") {
			cidrs = append(cidrs, cidr)
		}
		return cidrs, nil
	}

	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})

	if err != nil {
		log.Error().Err(err).Msgf("Fails to get node info of cluster")
		return nil, err
	}

	for _, node := range nodeList.Items {
		if node.Spec.PodCIDR != "" && len(node.Spec.PodCIDR) != 0 {
			cidrs = append(cidrs, node.Spec.PodCIDR)
		}
	}

	if len(cidrs) == 0 {
		samples, err2 := getPodCidrByInstance(ctx, clientset)
		if err2 != nil {
			return nil, err2
		}
		for _, sample := range samples.ToSlice() {
			cidrs = append(cidrs, fmt.Sprint(sample))
		}
	}

	return cidrs, nil
}

func getPodCidrByInstance(ctx context.Context, clientset kubernetes.Interface) (samples mapset.Set, err error) {
	log.Info().Msgf("Fail to get pod cidr from node.Spec.PODCIDR, try to get with pod sample")
	podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Fails to get service info of cluster")
		return
	}

	samples = mapset.NewSet()
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" && pod.Status.PodIP != "None" {
			samples.Add(getCidrFromSample(pod.Status.PodIP))
		}
	}
	return
}

func getServiceCidr(serviceList []coreV1.Service) (cidr []string, err error) {
	samples := mapset.NewSet()
	for _, service := range serviceList {
		if service.Spec.ClusterIP != "" && service.Spec.ClusterIP != "None" {
			samples.Add(getCidrFromSample(service.Spec.ClusterIP))
		}
	}

	for _, sample := range samples.ToSlice() {
		cidr = append(cidr, fmt.Sprint(sample))
	}
	return
}

func getCidrFromSample(sample string) string {
	return strings.Join(append(strings.Split(sample, ".")[:2], []string{"0", "0"}...), ".") + "/16"
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
		log.Info().Msgf("Waiting for shadow pod %s ...", podName)
	} else {
		log.Info().Msg("Waiting for shadow pod ...")
	}
}

func createService(metaAndSpec *SvcMetaAndSpec) *coreV1.Service {
	var servicePorts []coreV1.ServicePort
	metaAndSpec.Meta.Annotations[common.KTLastHeartBeat] = util.GetTimestamp()

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
	annotations[common.KTRefCount] = "1"
	annotations[common.KTLastHeartBeat] = util.GetTimestamp()
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
						Key:  common.SSHAuthKey,
						Path: "authorized_keys",
					},
				},
			},
		},
	}
	return sshVolume
}

func addTunHostPath(pod *coreV1.Pod) {
	path := "/dev/net/tun"

	pod.Spec.Volumes = append(pod.Spec.Volumes, coreV1.Volume{
		Name: "tun",
		VolumeSource: coreV1.VolumeSource{
			HostPath: &coreV1.HostPathVolumeSource{Path: path},
		},
	})

	for i := range pod.Spec.Containers {
		c := &pod.Spec.Containers[i]
		if c.Name != common.DefaultContainer {
			continue
		} else {
			c.VolumeMounts = append(c.VolumeMounts, coreV1.VolumeMount{
				Name:      "tun",
				MountPath: path,
			})
			c.SecurityContext.Capabilities.Add = append(c.SecurityContext.Capabilities.Add, "NET_ADMIN")
			break
		}
	}
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
