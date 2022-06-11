package birdseye

import (
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

func GetKtPodsAndAllServices(namespace string) ([]coreV1.Pod, []appV1.Deployment, []coreV1.Service, error) {
	pods, err := cluster.Ins().GetPodsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	apps, err := cluster.Ins().GetDeploymentsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	svcs, err := cluster.Ins().GetAllServiceInNamespace(opt.Get().Global.Namespace)
	if err != nil {
		return nil, nil, nil, err
	}
	return pods.Items, apps.Items, svcs.Items, nil
}

func ShowConnectors(pods []coreV1.Pod, apps []appV1.Deployment) {
	unknownUserCount := 0
	for _, pod := range pods {
		unknownUserCount += checkConnector(pod.Labels, pod.Annotations)
	}
	for _, app := range apps {
		unknownUserCount += checkConnector(app.Labels, app.Annotations)
	}
	if unknownUserCount > 0 {
		log.Info().Msgf("%d unknown users", unknownUserCount)
	}
}

func checkConnector(labels map[string]string, annotations map[string]string) int {
	if role, exists := labels[util.KtRole]; !exists || role != util.RoleConnectShadow {
		return 0
	}
	if user, exists := annotations[util.KtUser]; exists {
		lastHeartBeat := util.ParseTimestamp(annotations[util.KtLastHeartBeat])
		if lastHeartBeat > 0 {
			lastActiveInMin := (util.GetTime() - lastHeartBeat) / 60
			log.Info().Msgf("%s (last active %d min ago)", user, lastActiveInMin)
		} else {
			log.Info().Msgf("%s", user)
		}
	} else {
		return 1
	}
	return 0
}