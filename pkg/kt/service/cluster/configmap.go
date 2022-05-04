package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labelApi "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// GetConfigMap get configmap
func (k *Kubernetes) GetConfigMap(name, namespace string) (*coreV1.ConfigMap, error) {
	return k.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetConfigMapsByLabel get deployments by label
func (k *Kubernetes) GetConfigMapsByLabel(labels map[string]string, namespace string) (pods *coreV1.ConfigMapList, err error) {
	return k.Clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelApi.SelectorFromSet(labels).String(),
	})
}

// RemoveConfigMap remove ConfigMap instance
func (k *Kubernetes) RemoveConfigMap(name, namespace string) (err error) {
	deletePolicy := metav1.DeletePropagationBackground
	return k.Clientset.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func (k *Kubernetes) UpdateConfigMapHeartBeat(name, namespace string) {
	if _, err := k.Clientset.CoreV1().ConfigMaps(namespace).
		Patch(context.TODO(), name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{}); err != nil {
		if healthy, exists := LastHeartBeatStatus["configmap_" + name]; healthy || !exists {
			log.Warn().Err(err).Msgf("Failed to update heart beat of config map %s", name)
		} else {
			log.Debug().Err(err).Msgf("Config map %s heart beat interrupted", name)
		}
		LastHeartBeatStatus["configmap_" + name] = false
	} else {
		log.Debug().Msgf("Heartbeat configmap %s ticked at %s", name, util.FormattedTime())
		LastHeartBeatStatus["configmap_" + name] = true
	}
}

func (k *Kubernetes) createConfigMapWithSshKey(labels map[string]string, sshcm string, namespace string,
	generator *util.SSHGenerator) (configMap *coreV1.ConfigMap, err error) {
	SetupHeartBeat(sshcm, namespace, k.UpdateConfigMapHeartBeat)

	labels = util.MergeMap(labels, map[string]string{util.ControlBy: util.KubernetesToolkit})
	return k.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &coreV1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sshcm,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{util.KtLastHeartBeat: util.GetTimestamp()},
		},
		Data: map[string]string{
			util.SshAuthKey:        string(generator.PublicKey),
			util.SshAuthPrivateKey: string(generator.PrivateKey),
		},
	}, metav1.CreateOptions{})
}
