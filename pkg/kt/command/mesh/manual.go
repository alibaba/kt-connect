package mesh

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
)

func ManualMesh(resourceName string) error {
	app, err := general.GetDeploymentByResourceName(resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}

	meshKey, meshVersion := getVersion(opt.Get().MeshOptions.VersionMark)
	shadowPodName := app.Name + common.MeshPodInfix + meshVersion
	labels := getMeshLabels(meshKey, meshVersion, app)
	annotations := make(map[string]string)
	if err = general.CreateShadowAndInbound(shadowPodName, opt.Get().MeshOptions.Expose, labels, annotations); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf(" Now you can update Istio rule by label '%s=%s' ", meshKey, meshVersion)
	log.Info().Msg("---------------------------------------------------------")
	return nil
}

func getMeshLabels(meshKey, meshVersion string, app *appV1.Deployment) map[string]string {
	labels := map[string]string{}
	if app != nil {
		for k, v := range app.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	labels[common.KtRole] = common.RoleMeshShadow
	labels[meshKey] = meshVersion
	return labels
}
