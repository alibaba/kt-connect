package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	"strings"
)

// GetKtPods fetch all shadow and router pods
func GetKtPods(ctx context.Context, k KubernetesInterface, namespace string) ([]coreV1.Pod, error) {
	if pods, err := k.GetPods(ctx, map[string]string{common.ControlBy: common.KubernetesTool}, namespace); err != nil {
		return nil, err
	} else {
		return pods.Items, nil
	}
}

// GetOrCreateShadow create shadow
func GetOrCreateShadow(ctx context.Context, k KubernetesInterface, name string, options *options.DaemonOptions, labels, annotations, envs map[string]string) (
	podIP, podName, sshcm string, credential *util.SSHCredential, err error) {

	component := labels[common.KtComponent]
	identifier := strings.ToLower(util.RandomString(4))
	if options.ConnectOptions.ShareShadow {
		identifier = "shared"
	}
	sshcm = fmt.Sprintf("kt-%s-public-key-%s", component, identifier)
	privateKeyPath := util.PrivateKeyPath(component, identifier)

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.WithLabels) {
		labels[k] = v
	}
	for k, v := range util.String2Map(options.WithAnnotations) {
		annotations[k] = v
	}
	annotations[common.KtUser] = util.GetLocalUserName()
	resourceMeta := ResourceMeta{
		Name:        name,
		Namespace:   options.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
	sshKeyMeta := SSHkeyMeta{
		SshConfigMapName: sshcm,
		PrivateKeyPath:   privateKeyPath,
	}

	if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
		pod, generator, err2 := tryGetExistingShadowRelatedObjs(ctx, k, &resourceMeta, &sshKeyMeta)
		if err2 != nil {
			err = err2
			return
		}
		if pod != nil && generator != nil {
			podIP, podName, credential = shadowResult(pod, generator)
			return
		}
	}

	podIP, podName, credential, err = createShadow(ctx, k, &PodMetaAndSpec{&resourceMeta, options.Image, envs},
		&sshKeyMeta, options)
	return
}

func createShadow(ctx context.Context, k KubernetesInterface, metaAndSpec *PodMetaAndSpec, sshKeyMeta *SSHkeyMeta, options *options.DaemonOptions) (
	podIP string, podName string, credential *util.SSHCredential, err error) {

	generator, err := util.Generate(sshKeyMeta.PrivateKeyPath)
	if err != nil {
		return
	}
	configMap, err2 := k.CreateConfigMapWithSshKey(ctx, metaAndSpec.Meta.Labels, sshKeyMeta.SshConfigMapName, metaAndSpec.Meta.Namespace, generator)

	if err2 != nil {
		err = err2
		return
	}
	log.Info().Msgf("Successful create config map %v", configMap.Name)

	pod, err2 := createAndGetPod(ctx, k, metaAndSpec, sshKeyMeta.SshConfigMapName, options)
	if err2 != nil {
		err = err2
		return
	}
	podIP, podName, credential = shadowResult(pod, generator)
	return
}

func createAndGetPod(ctx context.Context, k KubernetesInterface, metaAndSpec *PodMetaAndSpec, sshcm string, options *options.DaemonOptions) (*coreV1.Pod, error) {
	localIPAddress := util.GetOutboundIP()
	log.Debug().Msgf("Client address %s", localIPAddress)
	metaAndSpec.Meta.Labels[common.KtRemoteAddress] = localIPAddress
	metaAndSpec.Meta.Labels[common.KtName] = metaAndSpec.Meta.Name

	err := k.CreateShadowPod(ctx, metaAndSpec, sshcm, options)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Deploying shadow pod %s in namespace %s", metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)

	return k.WaitPodReady(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace)
}

func tryGetExistingShadowRelatedObjs(ctx context.Context, k KubernetesInterface, resourceMeta *ResourceMeta, sshKeyMeta *SSHkeyMeta) (pod *coreV1.Pod, generator *util.SSHGenerator, err error) {
	_, shadowError := k.GetPod(ctx, resourceMeta.Name, resourceMeta.Namespace)
	if shadowError != nil {
		return
	}

	configMap, configMapError := k.GetConfigMap(ctx, sshKeyMeta.SshConfigMapName, resourceMeta.Namespace)
	if configMapError != nil {
		err = errors.New("Found shadow pod but no configMap. Please delete the pod " + resourceMeta.Name)
		return
	}

	generator = util.NewSSHGenerator(configMap.Data[common.SshAuthPrivateKey], configMap.Data[common.SshAuthKey], sshKeyMeta.PrivateKeyPath)

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[common.SshAuthPrivateKey]))
	if err != nil {
		return
	}

	return getShadowPod(ctx, k, resourceMeta, generator)
}

func getShadowPod(ctx context.Context, k KubernetesInterface, resourceMeta *ResourceMeta, generator *util.SSHGenerator) (pod *coreV1.Pod, sshGenerator *util.SSHGenerator, err error) {
	podList, err := k.GetPods(ctx, resourceMeta.Labels, resourceMeta.Namespace)
	if err != nil {
		return
	}
	if len(podList.Items) == 1 {
		log.Info().Msgf("Found shadow daemon pod, reuse it")
		if err = k.IncreaseRef(ctx, resourceMeta.Name, resourceMeta.Namespace); err != nil {
			return
		}
		return &(podList.Items[0]), generator, nil
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
