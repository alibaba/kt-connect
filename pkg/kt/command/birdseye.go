package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/birdseye"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
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
	err := showServiceStatus()
	if err != nil {
		return err
	}

	if opt.Get().Birdseye.ShowConnector {
		err = showConnectors()
		if err != nil {
			return err
		}
	}
	return nil
}

func showServiceStatus() error {
	ktPods, ktSvcs, svcs, err := birdseye.GetKtPodsAndAllServices(opt.Get().Global.Namespace)
	if err != nil {
		return err
	}

	// service-name, service-description
	allServices := birdseye.GetServiceStatus(ktSvcs, ktPods, svcs)
	if opt.Get().Birdseye.SortBy == util.SortByName {
		birdseye.SortServiceArray(allServices, 0)
	} else if opt.Get().Birdseye.SortBy == util.SortByStatus {
		birdseye.SortServiceArray(allServices, 1)
	} else {
		return fmt.Errorf("invalid sort method: %s", opt.Get().Birdseye.SortBy)
	}
	log.Info().Msgf("---- Service in namespace %s ----", opt.Get().Global.Namespace)
	for _, svc := range allServices {
		log.Info().Msgf("> %s - %s", svc[0], svc[1])
	}
	return nil
}

func showConnectors() error {
	pods, apps, err := birdseye.GetKtPodsAndDeployments()
	if err != nil {
		return err
	}

	unknownUserCount := 0
	users := birdseye.GetConnectors(pods, apps)
	log.Info().Msgf("---- User connecting to cluster ----")
	for _, user := range users {
		if user == birdseye.UnknownUser {
			unknownUserCount++
		} else {
			log.Info().Msgf("> %s", user)
		}
	}
	if unknownUserCount > 0 {
		log.Info().Msgf("%d users in total (including %d unknown users)",
			len(users)+unknownUserCount, unknownUserCount)
	} else {
		log.Info().Msgf("%d users in total", len(users))
	}
	return nil
}
