package cluster

import (
	"context"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateRouterPod create router pod
func (k *Kubernetes) CreateRouterPod(name string, labels, annotations map[string]string, ports map[int]int) (*coreV1.Pod, error) {
	targetPorts := make([]int, 0)
	for _, v := range ports {
		targetPorts = append(targetPorts, v)
	}
	metaAndSpec := &PodMetaAndSpec{&ResourceMeta{
		Name:        name,
		Namespace:   opt.Get().Namespace,
		Labels:      labels,
		Annotations: annotations,
	}, opt.Get().MeshOptions.RouterImage, map[string]string{}, targetPorts, true}
	pod := createPod(metaAndSpec)
	if _, err := k.Clientset.CoreV1().Pods(metaAndSpec.Meta.Namespace).
		Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		return nil, err
	}
	SetupHeartBeat(metaAndSpec.Meta.Name, metaAndSpec.Meta.Namespace, k.UpdatePodHeartBeat)
	log.Info().Msgf("Router pod %s created", name)
	return k.WaitPodReady(name, opt.Get().Namespace, opt.Get().PodCreationWaitTime)
}