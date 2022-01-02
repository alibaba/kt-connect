package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
)

func CreateRouterPod(ctx context.Context, k KubernetesInterface, name string, options *options.DaemonOptions,
	labels, annotations map[string]string) error {
	metaAndSpec := PodMetaAndSpec{&ResourceMeta{
		Name:        name,
		Namespace:   options.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}, options.MeshOptions.RouterImage, map[string]string{}}
	if err := k.CreatePod(ctx, &metaAndSpec, options); err != nil {
		return err
	}
	log.Info().Msgf("Router pod %s created", name)
	if _, err := k.WaitPodReady(ctx, name, options.Namespace); err != nil {
		return err
	}
	return nil
}
