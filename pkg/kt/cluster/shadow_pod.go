package cluster

import (
	"context"
	"errors"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// GetKtResources fetch all kt pods and deployments
func GetKtResources(ctx context.Context, k KubernetesInterface, namespace string) ([]coreV1.Pod, []coreV1.ConfigMap, []coreV1.Service, error) {
	pods, err := k.GetPodsByLabel(ctx, map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	configmaps, err := k.GetConfigMapsByLabel(ctx, map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	services, err := k.GetServicesByLabel(ctx, map[string]string{common.ControlBy: common.KubernetesTool}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	return pods.Items, configmaps.Items, services.Items, nil
}

// GetOrCreateShadow create shadow
func GetOrCreateShadow(ctx context.Context, k KubernetesInterface, name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (
	string, string, *util.SSHCredential, error) {

	// record context data
	options.RuntimeOptions.Shadow = name

	// extra labels must be applied after origin labels
	for key, val := range util.String2Map(options.WithLabels) {
		labels[key] = val
	}
	for key, val := range util.String2Map(options.WithAnnotations) {
		annotations[key] = val
	}
	annotations[common.KtUser] = util.GetLocalUserName()
	resourceMeta := ResourceMeta{
		Name:        name,
		Namespace:   options.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
	sshKeyMeta := SSHkeyMeta{
		SshConfigMapName: name,
		PrivateKeyPath:   util.PrivateKeyPath(name),
	}

	if options.RuntimeOptions.Component == common.ComponentConnect && options.ConnectOptions.SharedShadow {
		pod, generator, err2 := tryGetExistingShadowRelatedObjs(ctx, k, &resourceMeta, &sshKeyMeta)
		if err2 != nil {
			return "", "", nil, err2
		}
		if pod != nil && generator != nil {
			podIP, podName, credential := shadowResult(pod, generator)
			return podIP, podName, credential, nil
		}
	}

	podMeta := PodMetaAndSpec{
		Meta:  &resourceMeta,
		Image: options.Image,
		Envs:  envs,
	}
	return createShadow(ctx, k, &podMeta, &sshKeyMeta, options)
}

func createShadow(ctx context.Context, k KubernetesInterface, metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta, options *options.DaemonOptions) (
	podIP string, podName string, credential *util.SSHCredential, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}

	configMap, err := k.CreateConfigMapWithSshKey(ctx, metaAndSpec.Meta.Labels, sshKeyMeta.SshConfigMapName, metaAndSpec.Meta.Namespace, generator)
	if err != nil {
		return
	}
	log.Info().Msgf("Successful create config map %v", configMap.Name)

	pod, err := createAndGetPod(ctx, k, metaAndSpec, sshKeyMeta.SshConfigMapName, options)
	if err != nil {
		return
	}
	podIP, podName, credential = shadowResult(pod, generator)
	return
}

func createAndGetPod(ctx context.Context, k KubernetesInterface, metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) (*coreV1.Pod, error) {
	err := k.CreateShadowPod(ctx, metaAndSpec, sshcm, options)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Deploying shadow pod %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)

	return k.WaitPodReady(ctx, metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, options.PodCreationWaitTime)
}

func tryGetExistingShadowRelatedObjs(ctx context.Context, k KubernetesInterface, resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (*coreV1.Pod, *util.SSHGenerator, error) {
	pod, ignorableErr := k.GetPod(ctx, resourceMeta.Name, resourceMeta.Namespace);
	if ignorableErr != nil {
		return nil, nil, nil
	}

	configMap, err := k.GetConfigMap(ctx, sshKeyMeta.SshConfigMapName, resourceMeta.Namespace)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			if pod.DeletionTimestamp == nil {
				log.Error().Msgf("Found shadow Pod without ConfigMap. Please delete the pod '%s'", resourceMeta.Name)
			} else {
				_, err = k.WaitPodTerminate(ctx, resourceMeta.Name, resourceMeta.Namespace)
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

	pod, err = getShadowPod(ctx, k, resourceMeta)
	return pod, generator, err
}

func getShadowPod(ctx context.Context, k KubernetesInterface, resourceMeta *ResourceMeta) (pod *coreV1.Pod, err error) {
	podList, err := k.GetPodsByLabel(ctx, resourceMeta.Labels, resourceMeta.Namespace)
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shadow daemon pod, reuse it")
		if err = k.IncreaseRef(ctx, resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return
		}
		return &(podList.Items[0]), nil
	} else if len(podList.Items) > 1 {
		err = errors.New("Found more than one pod with name " + resourceMeta.Name + ", please make sure these is only one in namespace " + resourceMeta.Namespace)
	}
	return
}

func shadowResult(pod *coreV1.Pod, generator *util.SSHGenerator) (string, string, *util.SSHCredential) {
	podIP := pod.Status.PodIP
	podName := pod.Name
	credential := util.NewDefaultSSHCredential()
	credential.PrivateKeyPath = generator.PrivateKeyPath
	return podIP, podName, credential
}
