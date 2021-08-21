package cluster

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	mapset "github.com/deckarep/golang-set"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesClient(kubeConfig string) (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	return
}

func getPodCidrs(clientset kubernetes.Interface, podCIDR string) (cidrs []string, err error) {
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
		samples, err2 := getPodCidrByInstance(clientset)
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

func getPodCidrByInstance(clientset kubernetes.Interface) (samples mapset.Set, err error) {
	log.Info().Msgf("Fail to get pod cidr from node.Spec.PODCIDR, try to get with pod sample")
	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Error().Msg("Fails to get service info of cluster")
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

func getServiceCidr(serviceList []v1.Service) (cidr []string, err error) {
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

func getTargetPod(name string, labelsKeys []string, podList []*v1.Pod) *v1.Pod {
	for _, p := range podList {
		if len(p.Labels) <= 0 {
			// almost impossible
			continue
		}
		item, containKey := p.Labels[labelsKeys[0]]
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

func service(name, namespace string, labels map[string]string, external bool, port int) *v1.Service {
	var ports []v1.ServicePort
	annotations := map[string]string{common.KTLastHeartBeat: util.GetTimestamp()}

	ports = append(ports, v1.ServicePort{
		Name:       name,
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
	})

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}
	if external {
		service.Spec.Type = v1.ServiceTypeLoadBalancer
	}
	return service
}

func container(image string, args []string, envs map[string]string, options *options.DaemonOptions) v1.Container {
	var envVar []v1.EnvVar
	for k, v := range envs {
		envVar = append(envVar, v1.EnvVar{Name: k, Value: v})
	}
	var pullPolicy v1.PullPolicy
	if options.ForceUpdateShadow {
		pullPolicy = "Always"
	} else {
		pullPolicy = "IfNotPresent"
	}
	return v1.Container{
		Name:            "standalone",
		Image:           image,
		ImagePullPolicy: pullPolicy,
		Args:            args,
		Env:             envVar,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "ssh-public-key",
				MountPath: fmt.Sprintf("/root/%s", common.SSHAuthKey),
			},
		},
		SecurityContext: &v1.SecurityContext{
			Capabilities: &v1.Capabilities{
				Add: []v1.Capability{
					"AUDIT_WRITE",
				},
			},
		},
	}
}

func deployment(metaAndSpec *PodMetaAndSpec, volume string, options *options.DaemonOptions) *appV1.Deployment {
	var args []string
	namespace := metaAndSpec.Meta.Namespace
	name := metaAndSpec.Meta.Name
	labels := metaAndSpec.Meta.Labels
	annotations := metaAndSpec.Meta.Annotations
	annotations[common.KTRefCount] = "1"
	annotations[common.KTLastHeartBeat] = util.GetTimestamp()
	image := metaAndSpec.Image
	envs := metaAndSpec.Envs
	dep := &appV1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: appV1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					ServiceAccountName: options.ServiceAccount,
					Containers: []v1.Container{
						container(image, args, envs, options),
					},
					Volumes: []v1.Volume{
						getSSHVolume(volume),
					},
				},
			},
		},
	}

	if options.ConnectOptions != nil && options.ConnectOptions.Method == common.ConnectMethodTun {
		addTunHostPath(dep)
	}

	return dep
}

func getSSHVolume(volume string) v1.Volume {
	sshVolume := v1.Volume{
		Name: "ssh-public-key",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: volume,
				},
				Items: []v1.KeyToPath{
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

func addTunHostPath(dep *appV1.Deployment) {
	path := "/dev/net/tun"

	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "tun",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{Path: path},
		},
	})

	for i := range dep.Spec.Template.Spec.Containers {
		c := &dep.Spec.Template.Spec.Containers[i]
		if c.Name != "standalone" {
			continue
		} else {
			c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
				Name:      "tun",
				MountPath: path,
			})

			c.SecurityContext.Capabilities.Add = append(c.SecurityContext.Capabilities.Add, "NET_ADMIN")
			break
		}
	}
}
