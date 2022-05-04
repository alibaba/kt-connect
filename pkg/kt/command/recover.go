package command

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/command/recover"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

// NewRecoverCommand return new recover command
func NewRecoverCommand(action ActionInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "recover",
		Short: "Restore traffic of specified kubernetes service changed by exchange or mesh",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opt.Get().SkipTimeDiff = true
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("name of service to recover is required")
			}
			return action.Recover(args[0])
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl recover [command options]"))
	cmd.Long = cmd.Short

	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	return cmd
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

	if originSelector, exists := svc.Annotations[util.KtSelector]; exists {
		var selector map[string]string
		if err = json.Unmarshal([]byte(originSelector), &selector); err != nil {
			return fmt.Errorf("service %s has %s annotation, but selecting nothing", serviceName, util.KtSelector)
		}
		log.Debug().Msgf("Recovering selector to %v", selector)
		svc.Spec.Selector = selector
		delete(svc.Annotations, util.KtSelector)
		if targetRole == util.RoleRouter {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recover.HandleMeshedByAutoService(svc, targetDeployment, targetPod)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recover.HandleExchangedBySelectorService(svc, targetDeployment, targetPod)
		} else {
			log.Info().Msgf("Service %s is selecting non-kt pods, recovering", serviceName)
			return recover.HandleServiceSelectorAndRemotePods(svc, targetDeployment, targetPod)
		}
	} else {
		if targetRole == util.RoleMeshShadow {
			log.Info().Msgf("Service %s is meshed, recovering", serviceName)
			return recover.HandleMeshedByManualService(svc, targetDeployment, targetPod)
		} else if targetRole == util.RoleExchangeShadow {
			log.Info().Msgf("Service %s is exchanged, recovering", serviceName)
			return recover.HandleExchangedByScaleService(svc, targetDeployment, targetPod)
		} else if needUnlock {
			return recover.UnlockServiceOnly(svc)
		}
	}
	log.Info().Msgf("Service %s neither exchanged nor meshed by kt, nothing would be done", serviceName)
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

func checkAndMarkUnlock(serviceName string, svc *coreV1.Service) bool {
	if _, exists := svc.Annotations[util.KtLock]; exists {
		log.Info().Msgf("Unlocking service %s", serviceName)
		delete(svc.Annotations, util.KtLock)
		return true
	}
	return false
}