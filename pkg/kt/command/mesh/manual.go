package mesh

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
)

func ManualMesh(ctx context.Context, k cluster.KubernetesInterface, deploymentName string, opts *options.DaemonOptions) error {
	app, err := k.GetDeployment(ctx, deploymentName, opts.Namespace)
	if err != nil {
		return err
	}

	meshKey, meshVersion := getVersion(opts.MeshOptions.VersionMark)
	shadowPodName := deploymentName + common.MeshPodInfix + meshVersion
	labels := getMeshLabels(shadowPodName, meshKey, meshVersion, app)
	annotations := make(map[string]string)
	if err = createShadowAndInbound(ctx, k, shadowPodName, labels, annotations, opts); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf(" Now you can update Istio rule by label '%s=%s' ", meshKey, meshVersion)
	log.Info().Msg("---------------------------------------------------------")
	return nil
}
