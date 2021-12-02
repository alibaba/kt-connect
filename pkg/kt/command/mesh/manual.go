package mesh

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
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

func getMeshLabels(workload, meshKey, meshVersion string, app *appV1.Deployment) map[string]string {
	labels := map[string]string{}
	if app != nil {
		for k, v := range app.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	labels[common.KtComponent] = common.ComponentMesh
	labels[common.KtName] = workload
	labels[meshKey] = meshVersion
	return labels
}
