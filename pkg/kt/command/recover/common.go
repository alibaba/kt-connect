package recover

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)


func UnlockServiceOnly(svc *coreV1.Service) error {
	_, err := cluster.Ins().UpdateService(svc)
	return err
}

func HandleExchangedByScaleService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	if _, err := cluster.Ins().UpdateService(svc); err != nil {
		return err
	}
	config := make(map[string]string)
	if pod != nil && pod.Annotations != nil {
		config = util.String2Map(pod.Annotations[util.KtConfig])
		log.Info().Msgf("Deleting shadow pod %s", pod.Name)
		_ = cluster.Ins().RemovePod(pod.Name, pod.Namespace)
	}
	if len(config) == 0 && deployment != nil && deployment.Annotations != nil {
		config = util.String2Map(deployment.Annotations[util.KtConfig])
		log.Info().Msgf("Deleting shadow deployment %s", deployment.Name)
		_ = cluster.Ins().RemoveDeployment(deployment.Name, deployment.Namespace)
	}
	replica, _ := strconv.ParseInt(config["replicas"], 10, 32)
	app := config["app"]
	if replica > 0 && app != "" {
		originReplica := int32(replica)
		return cluster.Ins().ScaleTo(app, svc.Namespace, &originReplica)
	}
	return nil
}

func HandleMeshedByManualService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	return HandleServiceSelectorAndRemotePods(svc, deployment, pod)
}

func HandleExchangedBySelectorService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	return HandleServiceSelectorAndRemotePods(svc, deployment, pod)
}

func HandleMeshedByAutoService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	// shadow pods, shadow deployments, shadow services
	if deployment != nil {
		return fmt.Errorf("service '%s' is meshed but selecting more than a router pod, cannot auto recover", svc.Name)
	} else if pod == nil {
		return fmt.Errorf("service '%s' is meshed without selecting a router pod, cannot auto recover", svc.Name)
	}
	// must delete router pod first, to avoid origin service recover by mesh watcher
	log.Info().Msgf("Deleting route pod %s", pod.Name)
	if err := cluster.Ins().RemovePod(pod.Name, pod.Namespace); err != nil {
		log.Debug().Err(err).Msgf("Failed to remove pod %s", pod.Name)
	}
	time.Sleep(1 * time.Second)
	if _, err := cluster.Ins().UpdateService(svc); err != nil {
		return err
	}
	log.Info().Msgf("Deleting stuntman service %s", svc.Name + util.StuntmanServiceSuffix)
	if err := cluster.Ins().RemoveService(svc.Name + util.StuntmanServiceSuffix, svc.Namespace); err != nil {
		log.Debug().Err(err).Msgf("Failed to remove service %s", svc.Name)
	}
	shadowLabels := map[string]string{
		util.ControlBy: util.KubernetesToolkit,
		util.KtRole:    util.RoleMeshShadow,
	}
	shadowSvcNames := make([]string, 0)
	if apps, err := cluster.Ins().GetDeploymentsByLabel(shadowLabels, svc.Namespace); err == nil {
		for _, shadowApp := range apps.Items {
			if strings.HasPrefix(shadowApp.Name, svc.Name + util.MeshPodInfix) {
				log.Info().Msgf("Deleting shadow deployment %s", shadowApp.Name)
				if err2 := cluster.Ins().RemoveDeployment(shadowApp.Name, shadowApp.Namespace); err2 != nil {
					log.Debug().Err(err2).Msgf("Failed to remove deployment %s", shadowApp.Name)
				}
				shadowSvcNames = append(shadowSvcNames, shadowApp.Name)
			}
		}
	}
	if pods, err := cluster.Ins().GetPodsByLabel(shadowLabels, svc.Namespace); err == nil {
		for _, shadowPod := range pods.Items {
			if strings.HasPrefix(shadowPod.Name, svc.Name + util.MeshPodInfix) && shadowPod.DeletionTimestamp == nil {
				log.Info().Msgf("Deleting shadow pod %s", shadowPod.Name)
				if err2 := cluster.Ins().RemovePod(shadowPod.Name, shadowPod.Namespace); err2 != nil {
					log.Debug().Err(err2).Msgf("Failed to remove pod %s", pod.Name)
				}
				shadowSvcNames = append(shadowSvcNames, shadowPod.Name)
			}
		}
	}
	for _, shadowSvc := range shadowSvcNames {
		log.Info().Msgf("Deleting shadow service %s", shadowSvc)
		if err := cluster.Ins().RemoveService(shadowSvc, svc.Namespace); err != nil {
			log.Debug().Err(err).Msgf("Failed to remove service %s", svc.Name)
		}
	}
	return nil
}

func HandleServiceSelectorAndRemotePods(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	if _, err := cluster.Ins().UpdateService(svc); err != nil {
		return err
	}
	if deployment != nil {
		log.Info().Msgf("Deleting shadow deployment %s", deployment.Name)
		_ = cluster.Ins().RemoveDeployment(deployment.Name, deployment.Namespace)
	}
	if pod != nil {
		log.Info().Msgf("Deleting shadow pod %s", pod.Name)
		_ = cluster.Ins().RemovePod(pod.Name, pod.Namespace)
	}
	return nil
}

