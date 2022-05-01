package mesh

import (
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
)

func ManualMesh(svc *coreV1.Service) error {
	meshKey, meshVersion := getVersion(opt.Get().MeshOptions.VersionMark)
	shadowPodName := svc.Name + util.MeshPodInfix + meshVersion
	labels := getMeshLabels(meshKey, meshVersion, svc)
	annotations := make(map[string]string)
	if err := general.CreateShadowAndInbound(shadowPodName, opt.Get().MeshOptions.Expose, labels,
		annotations, general.GetTargetPorts(svc)); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf(" Now you can update Istio rule by label '%s=%s' ", meshKey, meshVersion)
	log.Info().Msg("---------------------------------------------------------")
	return nil
}

func getMeshLabels(meshKey, meshVersion string, svc *coreV1.Service) map[string]string {
	labels := map[string]string{}
	if svc != nil {
		for k, v := range svc.Spec.Selector {
			labels[k] = v
		}
	}
	labels[util.KtRole] = util.RoleMeshShadow
	labels[meshKey] = meshVersion
	return labels
}
