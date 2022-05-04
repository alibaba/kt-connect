package cluster

import (
	"context"
	"encoding/base64"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddEphemeralContainer add ephemeral container to specified pod
func (k *Kubernetes) AddEphemeralContainer(containerName string, name string,
	envs map[string]string) (string, error) {
	pod, err := k.GetPod(name, opt.Get().Namespace)
	if err != nil {
		return "", err
	}

	privateKeyPath := util.PrivateKeyPath(name)
	generator, err := util.Generate(privateKeyPath)
	if err != nil {
		return "", err
	}
	configMap, err2 := k.createConfigMapWithSshKey(map[string]string{}, name, opt.Get().Namespace, generator)

	if err2 != nil {
		return "", fmt.Errorf("found shadow pod but no configMap. Please delete the pod %s", pod.Name)
	}

	err = util.WritePrivateKey(generator.PrivateKeyPath, []byte(configMap.Data[util.SshAuthPrivateKey]))

	privateKey := base64.StdEncoding.EncodeToString([]byte(configMap.Data[util.SshAuthPrivateKey]))

	ec := coreV1.EphemeralContainer{
		EphemeralContainerCommon: coreV1.EphemeralContainerCommon{
			Name:  containerName,
			Image: fmt.Sprintf("%s:v%s", util.ImageKtNavigator, opt.Get().RuntimeStore.Version),
			Env: []coreV1.EnvVar{
				{Name: util.SshAuthPrivateKey, Value: privateKey},
			},
			SecurityContext: &coreV1.SecurityContext{
				Capabilities: &coreV1.Capabilities{Add: []coreV1.Capability{"NET_ADMIN"}},
			},
		},
	}

	for k, v := range envs {
		ec.Env = append(ec.Env, coreV1.EnvVar{Name: k, Value: v})
	}

	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ec)

	pod, err = k.Clientset.CoreV1().Pods(pod.Namespace).UpdateEphemeralContainers(context.TODO(), pod.Name, pod, metav1.UpdateOptions{})
	return privateKeyPath, err
}

// RemoveEphemeralContainer remove ephemeral container from specified pod
func (k *Kubernetes) RemoveEphemeralContainer(_, podName string, namespace string) (err error) {
	// TODO: implement container removal
	return k.RemovePod(podName, namespace)
}
