package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/birdseye"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
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
	pods, apps, ktSvcs, svcs, err := birdseye.GetKtPodsAndAllServices(opt.Get().Global.Namespace)
	if err != nil {
		return err
	}

	if opt.Get().Birdseye.ShowConnector {
		log.Info().Msgf("---- User connecting to cluster ----")
		unknownUserCount := 0
		users := birdseye.GetConnectors(pods, apps)
		for _, user := range users {
			if user == birdseye.UnknownUser {
				unknownUserCount++
			} else {
				log.Info().Msgf("> %s", user)
			}
		}
		if unknownUserCount > 0 {
			log.Info().Msgf("%d users in total (including %d unknown users)",
				len(users) + unknownUserCount, unknownUserCount)
		} else {
			log.Info().Msgf("%d users in total", len(users))
		}
	}

	// service-name, service-description
	allServices := birdseye.GetServiceStatus(ktSvcs, pods, svcs)
	log.Info().Msgf("---- Service in namespace %s ----", opt.Get().Global.Namespace)
	for _, svc := range allServices {
		log.Info().Msgf("> %s - %s", svc[0], svc[1])
	}
	return nil
}
