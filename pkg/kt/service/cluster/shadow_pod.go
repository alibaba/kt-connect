package cluster

import (
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// GetOrCreateShadow create shadow
func (k *Kubernetes) GetOrCreateShadow(name string, labels, annotations, envs map[string]string, exposePorts string) (
	string, string, string, error) {
	// record context data
	opt.Get().RuntimeStore.Shadow = name

	// extra labels must be applied after origin labels
	for key, val := range util.String2Map(opt.Get().WithLabels) {
		labels[key] = val
	}
	for key, val := range util.String2Map(opt.Get().WithAnnotations) {
		annotations[key] = val
	}
	annotations[util.KtUser] = util.GetLocalUserName()
	resourceMeta := ResourceMeta{
		Name:        name,
		Namespace:   opt.Get().Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
	sshKeyMeta := SSHkeyMeta{
		SshConfigMapName: name,
		PrivateKeyPath:   util.PrivateKeyPath(name),
	}

	ports := make([]int, 0)
	if exposePorts != "" {
		portPairs := strings.Split(exposePorts, ",")
		for _, exposePort := range portPairs {
			_, port, err := util.ParsePortMapping(exposePort)
			if err != nil {
				log.Warn().Err(err).Msgf("invalid port")
			} else {
				ports = append(ports, port)
			}
		}
	}

	if opt.Get().RuntimeStore.Component == util.ComponentConnect && opt.Get().ConnectOptions.SharedShadow {
		pod, generator, err2 := k.tryGetExistingShadowRelatedObjs(&resourceMeta, &sshKeyMeta)
		if err2 != nil {
			return "", "", "", err2
		}
		if pod != nil && generator != nil {
			podIP, podName, credential := shadowResult(pod, generator)
			return podIP, podName, credential, nil
		}
	}

	podMeta := PodMetaAndSpec{
		Meta:  &resourceMeta,
		Image: opt.Get().Image,
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
	podIP, podName, privateKeyPath = shadowResult(pod, generator)
	return
}

func (k *Kubernetes) createAndGetPod(metaAndSpec *PodMetaAndSpec, sshcm string) (*coreV1.Pod, error) {
	err := k.createShadowPod(metaAndSpec, sshcm)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Deploying shadow pod %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)

	return k.WaitPodReady(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, opt.Get().PodCreationWaitTime)
}

// createShadowPod create shadow pod
func (k *Kubernetes) createShadowPod(metaAndSpec *PodMetaAndSpec, sshcm string) error {
	pod := createPod(metaAndSpec)
	pod.Spec.Containers[0].VolumeMounts = []coreV1.VolumeMount{
		{
			Name:      "ssh-public-key",
			MountPath: fmt.Sprintf("/root/%s", util.SshAuthKey),
		},
	}
	pod.Spec.Volumes = []coreV1.Volume{
		getSSHVolume(sshcm),
	}
	if _, err := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdatePodHeartBeat)
	return nil
}

func (k *Kubernetes) tryGetExistingShadowRelatedObjs(resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (*coreV1.Pod, *util.SSHGenerator, error) {
	pod, ignorableErr := k.GetPod(resourceMeta.Name, resourceMeta.Namespace);
	if ignorableErr != nil {
		return nil, nil, nil
	}

	configMap, err := k.GetConfigMap(sshKeyMeta.SshConfigMapName, resourceMeta.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			if pod.DeletionTimestamp == nil {
				log.Error().Msgf("Found shadow Pod without ConfigMap. Please delete the pod '%s'", resourceMeta.Name)
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

	pod, err = k.getShadowPod(resourceMeta)
	return pod, generator, err
}

func (k *Kubernetes) getShadowPod(resourceMeta *ResourceMeta) (pod *coreV1.Pod, err error) {
	podList, err := k.GetPodsByLabel(resourceMeta.Labels, resourceMeta.Namespace)
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shadow daemon pod, reuse it")
		if err = k.IncreaseRef(resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return
		}
		return &(podList.Items[0]), nil
	} else if len(podList.Items) > 1 {
		err = fmt.Errorf("found more than one pod with name %s, please make sure these is only one in namespace %s",
			resourceMeta.Name, resourceMeta.Namespace)
	}
	return
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

func shadowResult(pod *coreV1.Pod, generator *util.SSHGenerator) (string, string, string) {
	podIP := pod.Status.PodIP
	podName := pod.Name
	return podIP, podName, generator.PrivateKeyPath
}
