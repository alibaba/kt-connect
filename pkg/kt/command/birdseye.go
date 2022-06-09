package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

// NewBirdseyeCommand show a summary of cluster service network
func NewBirdseyeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "birdseye",
		Short: "Show summary of services status in cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("too many options specified (%s)", strings.Join(args, ",") )
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Birdseye()
		},
		Example: "ktctl birdseye [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(false))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Birdseye, opt.BirdseyeFlags())
	return cmd
}

func Birdseye() error {
	if opt.Get().Birdseye.ShowConnector {
		pods, err := cluster.Ins().GetPodsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit},
			opt.Get().Global.Namespace)
		if err != nil {
			return err
		}
		unknownUserCount := 0
		for _, pod := range pods.Items {
			if role, exists := pod.Labels[util.KtRole]; !exists || role != util.RoleConnectShadow {
				continue
			}
			if user, exists := pod.Annotations[util.KtUser]; exists {
				lastHeartBeat := util.ParseTimestamp(pod.Annotations[util.KtLastHeartBeat])
				if lastHeartBeat > 0 {
					lastActiveInMin := (util.GetTime() - lastHeartBeat) / 60
					log.Info().Msgf("%s (last active %d min ago)", user, lastActiveInMin)
				} else {
					log.Info().Msgf("%s", user)
				}
			} else {
				unknownUserCount++
			}
		}
		if unknownUserCount > 0 {
			log.Info().Msgf("%d unknown users", unknownUserCount)
		}
	}
	svcs, err := cluster.Ins().GetAllServiceInNamespace(opt.Get().Global.Namespace)
	if err != nil {
		return err
	}
	allServices := make([]string, 0)
	//exchangedServices := make([]string, 0)
	//meshedServices := make([]string, 0)
	//previewingServices := make([]string, 0)
	for _, svc := range svcs.Items {
		if cb, exists := svc.Annotations[util.ControlBy]; exists && cb == util.KubernetesToolkit {

		} else if !opt.Get().Birdseye.HideNaturalService {
			allServices = append(allServices, svc.Name)
		}
	}
	for _, svc := range allServices {
		log.Info().Msgf("%s", svc)
	}
	return nil
}
