package command

import (
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
	if svc.Annotations == nil {
		log.Info().Msgf("Service %s is clean and tidy, nothing would be done", serviceName)
		return nil
	}

	needUnlock := false
	if _, ok := svc.Annotations[util.KtLock]; ok {
		log.Info().Msgf("Unlocking service %s", serviceName)
		delete(svc.Annotations, util.KtLock)
		needUnlock = true
	}

	apps, err2 := cluster.Ins().GetDeploymentsByLabel(svc.Spec.Selector, opt.Get().Namespace)
	if err2 != nil {
		return err2
	}
	pods, err2 := cluster.Ins().GetPodsByLabel(svc.Spec.Selector, opt.Get().Namespace)
	if err2 != nil {
		return err2
	}
	targetRole := fetchTargetRole(apps, pods)

	if _, ok := svc.Annotations[util.KtSelector]; ok {
		delete(svc.Annotations, util.KtSelector)
		if targetRole == util.RoleRouter {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recoverMeshedByAutoService(svc)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recoverExchangedBySelectorService(svc)
		} else {
			log.Info().Msgf("Service %s is selecting non-kt pods, recovering", serviceName)
			return recoverServiceSelectorOnly(svc)
		}
	} else {
		if targetRole == util.RoleMeshShadow {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recoverMeshedByManualService(svc)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recoverExchangedByScaleService(svc)
		} else if needUnlock {
			return unlockServiceOnly(svc)
		}
	}
	log.Info().Msgf("Service %s neither exchanged nor meshed by kt, nothing would be done", serviceName)
	return nil
}

func unlockServiceOnly(svc *coreV1.Service) error {
	return nil
}

func recoverExchangedByScaleService(svc *coreV1.Service) error {
	return nil
}

func recoverMeshedByManualService(svc *coreV1.Service) error {
	return nil
}

func recoverServiceSelectorOnly(svc *coreV1.Service) error {
	return nil
}

func recoverExchangedBySelectorService(svc *coreV1.Service) error {
	return nil
}

func recoverMeshedByAutoService(svc *coreV1.Service) error {
	return nil
}

func fetchTargetRole(apps *appV1.DeploymentList, pods *coreV1.PodList) string {
	if len(apps.Items) > 0 {
		for _, app := range apps.Items {
			if app.Annotations != nil {
				if role, ok2 := app.Annotations[util.KtRole]; ok2 {
					return role
				}
			}
		}
	} else if len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if pod.Annotations != nil {
				if role, ok2 := pod.Annotations[util.KtRole]; ok2 {
					return role
				}
			}
		}
	}
	return ""
}
