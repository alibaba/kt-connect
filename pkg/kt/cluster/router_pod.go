package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

func GetOrCreateRouterPod(ctx context.Context, k KubernetesInterface, name string, options *options.DaemonOptions) error {
	metaAndSpec := PodMetaAndSpec{&ResourceMeta{
		Name:        name,
		Namespace:   options.Namespace,
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}, options.MeshOptions.RouterImage, map[string]string{}}
	k.CreatePod(ctx, &metaAndSpec, options)
	return nil
}
