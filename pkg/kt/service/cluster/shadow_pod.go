package cluster

import (
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// GetOrCreateShadow create shadow pod or deployment
func (k *Kubernetes) GetOrCreateShadow(name string, labels, annotations, envs map[string]string, exposePorts string, portNameDict map[int]string) (
	string, string, string, error) {
	// record context data
	opt.Store.Shadow = name

	// extra labels must be applied after origin labels
	for key, val := range util.String2Map(opt.Get().Global.WithLabel) {
		labels[key] = val
	}
	for key, val := range util.String2Map(opt.Get().Global.WithAnnotation) {
		annotations[key] = val
	}
	annotations[util.KtUser] = util.GetLocalUserName()
	resourceMeta := ResourceMeta{
		Name:        name,
		Namespace:   opt.Get().Global.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
	sshKeyMeta := SSHkeyMeta{
		SshConfigMapName: name,
		PrivateKeyPath:   util.PrivateKeyPath(name),
	}

	ports := map[string]int{}
	if exposePorts != "" {
		portPairs := strings.Split(exposePorts, ",")
		for _, exposePort := range portPairs {
			_, port, err := util.ParsePortMapping(exposePort)
			if err != nil {
				log.Warn().Err(err).Msgf("invalid port")
			} else {
				// TODO: assume port using http protocol for istio constraint, should support user-defined protocol
				name = fmt.Sprintf("http-%d", port)
				if n, exists := portNameDict[port]; exists {
					name = n
				}
				ports[name] = port
			}
		}
	}

	if opt.Store.Component == util.ComponentConnect && opt.Get().Connect.ShareShadow {
		pod, generator, err2 := k.tryGetExistingShadows(&resourceMeta, &sshKeyMeta)
		if err2 != nil {
			return "", "", "", err2
		}
		if pod != nil && generator != nil {
			return pod.Status.PodIP, pod.Name, generator.PrivateKeyPath, nil
		}
	}

	podMeta := PodMetaAndSpec{
		Meta:  &resourceMeta,
		Image: opt.Get().Global.Image,
		Envs:  envs,
		Ports: ports,
	}
	return k.createShadow(&podMeta, &sshKeyMeta)
}

func (k *Kubernetes) createShadow(metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta) (
	podIP string, podName string, privateKeyPath string, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}

	configMap, err := k.createConfigMapWithSshKey(metaAndSpec.Meta.Labels, sshKeyMeta.SshConfigMapName, metaAndSpec.Meta.Namespace, generator)
	if err != nil {
		return
	}
	log.Info().Msgf("Successful create config map %v", configMap.Name)

	pod, err := k.createAndGetPod(metaAndSpec, sshKeyMeta.SshConfigMapName)
	if err != nil {
		return
	}
	return pod.Status.PodIP, pod.Name, generator.PrivateKeyPath, nil
}

func (k *Kubernetes) createAndGetPod(metaAndSpec *PodMetaAndSpec, sshcm string) (*coreV1.Pod, error) {
	if opt.Get().Global.UseShadowDeployment {
		if err := k.createShadowDeployment(metaAndSpec, sshcm); err != nil {
			return nil, err
		}
		log.Info().Msgf("Creating shadow deployment %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)
		delete(metaAndSpec.Meta.Labels, util.ControlBy)
		pods, err := k.WaitPodsReady(metaAndSpec.Meta.Labels, metaAndSpec.Meta.Namespace, opt.Get().Global.PodCreationTimeout)
		if err != nil {
			return nil, err
		}
		return &pods[0], nil
	} else {
		if err := k.createShadowPod(metaAndSpec, sshcm); err != nil {
			return nil, err
		}
		log.Info().Msgf("Deploying shadow pod %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)
		return k.WaitPodReady(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, opt.Get().Global.PodCreationTimeout)
	}
}

func filterRunningPods(pods []coreV1.Pod) []coreV1.Pod {
	runningPods := make([]coreV1.Pod, 0)
	for _, pod := range pods {
		if pod.Status.Phase == coreV1.PodRunning && pod.DeletionTimestamp == nil {
			runningPods = append(runningPods, pod)
		}
	}
	return runningPods
}

// createShadowDeployment create shadow deployment
func (k *Kubernetes) createShadowDeployment(metaAndSpec *PodMetaAndSpec, sshcm string) error {
	deployment := createDeployment(metaAndSpec)
	k.appendSshVolume(&deployment.Spec.Template.Spec, sshcm)
	if _, err := k.Clientset.AppsV1().Deployments(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), deployment, metav1.CreateOptions{}); err != nil {
		return err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdateDeploymentHeartBeat)
	return nil
}

// createShadowPod create shadow pod
func (k *Kubernetes) createShadowPod(metaAndSpec *PodMetaAndSpec, sshcm string) error {
	pod := createPod(metaAndSpec)
	k.appendSshVolume(&pod.Spec, sshcm)
	if _, err := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdatePodHeartBeat)
	return nil
}

func (k *Kubernetes) appendSshVolume(podSpec *coreV1.PodSpec, sshcm string) {
	podSpec.Containers[0].VolumeMounts = []coreV1.VolumeMount{
		{
			Name:      "ssh-public-key",
			MountPath: fmt.Sprintf("/root/%s", util.SshAuthKey),
		},
	}
	podSpec.Volumes = []coreV1.Volume{
		getSSHVolume(sshcm),
	}
}

func (k *Kubernetes) tryGetExistingShadows(resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (*coreV1.Pod, *util.SSHGenerator, error) {
	var app *appV1.Deployment
	var pod *coreV1.Pod
	if opt.Get().Global.UseShadowDeployment {
		app2, err := k.GetDeployment(resourceMeta.Name, resourceMeta.Namespace)
		if err != nil {
			// shared deployment not found is ok, return without error
			return nil, nil, nil
		}
		app = app2
		podList, err := k.GetPodsByLabel(app.Spec.Selector.MatchLabels, resourceMeta.Namespace)
		if err != nil || len(podList.Items) == 0 {
			log.Error().Err(err).Msgf("Found shadow deployment '%s' but cannot fetch it's pod", resourceMeta.Name)
			return nil, nil, err
		} else if len(podList.Items) > 1 {
			log.Warn().Msgf("Found more than one shadow pod with labels %v", app.Spec.Selector.MatchLabels)
			return nil, nil, err
		}
		pod = &podList.Items[0]
	} else {
		pod2, err := k.GetPod(resourceMeta.Name, resourceMeta.Namespace)
		if err != nil {
			// shared pod not found is ok, return without error
			return nil, nil, nil
		}
		pod = pod2
	}

	configMap, err := k.GetConfigMap(sshKeyMeta.SshConfigMapName, resourceMeta.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			if pod.DeletionTimestamp == nil {
				log.Error().Msgf("Found shadow pod without configmap. Please delete the pod '%s'", resourceMeta.Name)
			} else {
				_, err = k.WaitPodTerminate(resourceMeta.Name, resourceMeta.Namespace)
				if k8sErrors.IsNotFound(err) {
					// Pod already terminated
					return nil, nil, nil
				}
			}
		}
		return nil, nil, err
	}

	generator := util.NewSSHGenerator(configMap.Data[util.SshAuthPrivateKey], configMap.Data[util.SshAuthKey], sshKeyMeta.PrivateKeyPath)

	if err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[util.SshAuthPrivateKey])); err != nil {
		return nil, nil, err
	}

	if opt.Get().Global.UseShadowDeployment {
		log.Info().Msgf("Found shadow daemon deployment, reuse it")
		if err = k.IncreaseDeploymentRef(resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return nil, nil, err
		}
	} else {
		log.Info().Msgf("Found shadow daemon pod, reuse it")
		if err = k.IncreasePodRef(resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return nil, nil, err
		}
	}
	return pod, generator, nil
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
						Key:  util.SshAuthKey,
						Path: "authorized_keys",
					},
				},
			},
		},
	}
	return sshVolume
}

