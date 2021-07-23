package cluster

import (
	"fmt"
	"strconv"
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

func getPodCirds(clientset kubernetes.Interface, podCIDR string) (cidrs []string, err error) {
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

func getPodCirdByInstance(clientset kubernetes.Interface) (samples mapset.Set, err error) {
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

func getServiceCird(serviceList []v1.Service) (cidr []string, err error) {
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
	time.Sleep(time.Second)
	if len(podName) > 0 {
		log.Info().Msgf("Pod: %s is running, but not ready", podName)
	} else {
		log.Info().Msg("Shadow pods not ready ...")
	}
}

func service(name, namespace string, labels map[string]string, external bool, port int) *v1.Service {
	var ports []v1.ServicePort
	ports = append(ports, v1.ServicePort{
		Name:       name,
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
	})

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
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
	if options.Debug {
		log.Debug().Msg("Create shadow with debug mode")
		args = append(args, "--debug")
	}

	namespace := metaAndSpec.Meta.Namespace
	name := metaAndSpec.Meta.Name
	labels := metaAndSpec.Meta.Labels
	annotations := metaAndSpec.Meta.Annotations
	annotations[common.RefCount] = "1"
	annotations[common.KTLastHeartBeat] = strconv.FormatInt(time.Now().Unix(), 10)
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

	if options.ConnectOptions.Method == common.ConnectMethodTun {
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
