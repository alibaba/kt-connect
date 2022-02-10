package cluster

import (
	"context"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
)

func CreateRouterPod(ctx context.Context, name string,
	labels, annotations map[string]string) error {
	metaAndSpec := PodMetaAndSpec{&ResourceMeta{
		Name:        name,
		Namespace:   opt.Get().Namespace,
		Labels:      labels,
		Annotations: annotations,
	}, opt.Get().MeshOptions.RouterImage, map[string]string{}}
	if err := Ins().CreatePod(ctx, &metaAndSpec); err != nil {
		return err
	}
	log.Info().Msgf("Router pod %s created", name)
	if _, err := Ins().WaitPodReady(ctx, name, opt.Get().Namespace, opt.Get().PodCreationWaitTime); err != nil {
		return err
	}
	return nil
}
