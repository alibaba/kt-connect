package command

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
)

// NewRecoverCommand return new recover command
func NewRecoverCommand(action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "recover",
		Usage: "restore traffic of specified kubernetes service changed by exchange or mesh",
		UsageText: "ktctl recover [command options]",
		Flags: general.RecoverActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}

			if len(c.Args()) == 0 {
				return fmt.Errorf("name of service to recover is required")
			}

			return action.Recover(c.Args().First())
		},
	}
}

// Recover delete unavailing shadow pods
func (action *Action) Recover(serviceName string) error {
	svc, err := cluster.Ins().GetService(serviceName, opt.Get().Namespace)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to fetch service %s", serviceName)
	}

	apps, err := cluster.Ins().GetDeploymentsByLabel(svc.Spec.Selector, svc.Namespace)
	if err != nil {
		return err
	}
	pods, err := cluster.Ins().GetPodsByLabel(svc.Spec.Selector, svc.Namespace)
	if err != nil {
		return err
	}
	targetDeployment, targetPod, targetRole := fetchTargetRole(apps, pods)
	log.Debug().Msgf("Target role is: %s", targetRole)

	if svc.Annotations == nil {
		// put an empty map to avoid npe
		svc.Annotations = map[string]string{}
		if targetRole == "" {
			if svc.Spec.Selector[util.KtRole] != "" {
				log.Error().Msgf("Service %s is selecting kt pods, but cannot be recovered automatically", serviceName)
			} else {
				log.Info().Msgf("Service %s is clean and tidy, nothing would be done", serviceName)
			}
			return nil
		}
	}

	needUnlock := checkAndMarkUnlock(serviceName, svc)

	if originSelector, ok := svc.Annotations[util.KtSelector]; ok {
		var selector map[string]string
		if err = json.Unmarshal([]byte(originSelector), &selector); err != nil {
			return fmt.Errorf("service %s has %s annotation, but selecting nothing", serviceName, util.KtSelector)
		}
		log.Debug().Msgf("Recovering selector to %v", selector)
		svc.Spec.Selector = selector
		delete(svc.Annotations, util.KtSelector)
		if targetRole == util.RoleRouter {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recoverMeshedByAutoService(svc, targetDeployment, targetPod)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recoverExchangedBySelectorService(svc, targetDeployment, targetPod)
		} else {
			log.Info().Msgf("Service %s is selecting non-kt pods, recovering", serviceName)
			return recoverServiceSelectorAndRemotePods(svc, targetDeployment, targetPod)
		}
	} else {
		if targetRole == util.RoleMeshShadow {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recoverMeshedByManualService(svc, targetDeployment, targetPod)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recoverExchangedByScaleService(svc, targetDeployment, targetPod)
		} else if needUnlock {
			return unlockServiceOnly(svc)
		}
	}
	log.Info().Msgf("Service %s neither exchanged nor meshed by kt, nothing would be done", serviceName)
	return nil
}

func checkAndMarkUnlock(serviceName string, svc *coreV1.Service) bool {
	if _, ok := svc.Annotations[util.KtLock]; ok {
		log.Info().Msgf("Unlocking service %s", serviceName)
		delete(svc.Annotations, util.KtLock)
		return true
	}
	return false
}

func unlockServiceOnly(svc *coreV1.Service) error {
	_, err := cluster.Ins().UpdateService(svc)
	return err
}

func recoverExchangedByScaleService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
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

func recoverMeshedByManualService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	return recoverServiceSelectorAndRemotePods(svc, deployment, pod)
}

func recoverExchangedBySelectorService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
	return recoverServiceSelectorAndRemotePods(svc, deployment, pod)
}

func recoverMeshedByAutoService(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
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

func recoverServiceSelectorAndRemotePods(svc *coreV1.Service, deployment *appV1.Deployment, pod *coreV1.Pod) error {
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

func fetchTargetRole(apps *appV1.DeploymentList, pods *coreV1.PodList) (*appV1.Deployment, *coreV1.Pod, string) {
	if len(apps.Items) > 0 {
		for _, app := range apps.Items {
			if app.Labels != nil {
				if role, ok2 := app.Labels[util.KtRole]; ok2 {
					return &app, nil, role
				}
			}
		}
	} else if len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if pod.Labels != nil && pod.DeletionTimestamp == nil {
				if role, ok2 := pod.Labels[util.KtRole]; ok2 {
					return nil, &pod, role
				}
			}
		}
	}
	return nil, nil, ""
}
