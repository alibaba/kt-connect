package cluster

import (
	"errors"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// GetKtResources fetch all kt pods and deployments
func GetKtResources(namespace string) ([]coreV1.Pod, []coreV1.ConfigMap, []coreV1.Service, error) {
	pods, err := Ins().GetPodsByLabel(map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	configmaps, err := Ins().GetConfigMapsByLabel(map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	services, err := Ins().GetServicesByLabel(map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	return pods.Items, configmaps.Items, services.Items, nil
}

// GetOrCreateShadow create shadow
func GetOrCreateShadow(name string, labels, annotations, envs map[string]string) (
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
	annotations[common.KtUser] = util.GetLocalUserName()
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

	if opt.Get().RuntimeStore.Component == common.ComponentConnect && opt.Get().ConnectOptions.SharedShadow {
		pod, generator, err2 := tryGetExistingShadowRelatedObjs(&resourceMeta, &sshKeyMeta)
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
	}
	return createShadow(&podMeta, &sshKeyMeta)
}

func createShadow(metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta) (
	podIP string, podName string, privateKeyPath string, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}

	configMap, err := Ins().CreateConfigMapWithSshKey(metaAndSpec.Meta.Labels, sshKeyMeta.SshConfigMapName, metaAndSpec.Meta.Namespace, generator)
	if err != nil {
		return
	}
	log.Info().Msgf("Successful create config map %v", configMap.Name)

	pod, err := createAndGetPod(metaAndSpec, sshKeyMeta.SshConfigMapName)
	if err != nil {
		return
	}
	podIP, podName, privateKeyPath = shadowResult(pod, generator)
	return
}

func createAndGetPod(metaAndSpec *PodMetaAndSpec, sshcm string) (*coreV1.Pod, error) {
	err := Ins().CreateShadowPod(metaAndSpec, sshcm)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Deploying shadow pod %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)

	return Ins().WaitPodReady(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, opt.Get().PodCreationWaitTime)
}

func tryGetExistingShadowRelatedObjs(resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (*coreV1.Pod, *util.SSHGenerator, error) {
	pod, ignorableErr := Ins().GetPod(resourceMeta.Name, resourceMeta.Namespace);
	if ignorableErr != nil {
		return nil, nil, nil
	}

	configMap, err := Ins().GetConfigMap(sshKeyMeta.SshConfigMapName, resourceMeta.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			if pod.DeletionTimestamp == nil {
				log.Error().Msgf("Found shadow Pod without ConfigMap. Please delete the pod '%s'", resourceMeta.Name)
			} else {
				_, err = Ins().WaitPodTerminate(resourceMeta.Name, resourceMeta.Namespace)
				if k8sErrors.IsNotFound(err) {
					// Pod already terminated
					return nil, nil, nil
				}
			}
		}
		return nil, nil, err
	}

	generator := util.NewSSHGenerator(configMap.Data[common.SshAuthPrivateKey], configMap.Data[common.SshAuthKey], sshKeyMeta.PrivateKeyPath)

	if err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SshAuthPrivateKey])); err != nil {
		return nil, nil, err
	}

	pod, err = getShadowPod(resourceMeta)
	return pod, generator, err
}

func getShadowPod(resourceMeta *ResourceMeta) (pod *coreV1.Pod, err error) {
	podList, err := Ins().GetPodsByLabel(resourceMeta.Labels, resourceMeta.Namespace)
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shadow daemon pod, reuse it")
		if err = Ins().IncreaseRef(resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return
		}
		return &(podList.Items[0]), nil
	} else if len(podList.Items) > 1 {
		err = errors.New("Found more than one pod with name " + resourceMeta.Name + ", please make sure these is only one in namespace " + resourceMeta.Namespace)
	}
	return
}

func shadowResult(pod *coreV1.Pod, generator *util.SSHGenerator) (string, string, string) {
	podIP := pod.Status.PodIP
	podName := pod.Name
	return podIP, podName, generator.PrivateKeyPath
}
