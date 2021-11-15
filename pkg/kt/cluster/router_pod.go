package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

func GetOrCreateRouterPod(ctx context.Context, k KubernetesInterface, name string, options *options.DaemonOptions, labels map[string]string) error {
	metaAndSpec := PodMetaAndSpec{&ResourceMeta{
		Name:        name,
		Namespace:   options.Namespace,
		Labels:      labels,
		Annotations: map[string]string{},
	}, options.MeshOptions.RouterImage, map[string]string{}}
	if err := k.CreatePod(ctx, &metaAndSpec, options); err != nil {
		return err
	}
	return nil
}
