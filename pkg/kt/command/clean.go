package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewCleanCommand return new connect command
func NewCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "clean",
		Short: "Delete unavailing resources created by kt from kubernetes cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Clean()
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl clean [command options]"))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Clean, opt.CleanFlags())
	return cmd
}

// Clean delete unavailing shadow pods
func Clean() error {
	if resourceToClean, err := clean.CheckClusterResources(); err != nil {
		log.Warn().Err(err).Msgf("Failed to clean up cluster resources")
	} else {
		if isEmpty(resourceToClean) {
			log.Info().Msg("No unavailing kt resource found (^.^)YYa!!")
		} else {
			if opt.Get().Clean.DryRun {
				clean.PrintClusterResourcesToClean(resourceToClean)
			} else {
				clean.TidyClusterResources(resourceToClean)
			}
		}
	}
	if !opt.Get().Clean.DryRun {
		clean.TidyLocalResources()
	}
	return nil
}

func isEmpty(r *clean.ResourceToClean) bool {
	return len(r.PodsToDelete) == 0 &&
		len(r.ConfigMapsToDelete) == 0 &&
		len(r.DeploymentsToDelete) == 0 &&
		len(r.DeploymentsToScale) == 0 &&
		len(r.ServicesToDelete) == 0 &&
		len(r.ServicesToUnlock) == 0 &&
		len(r.ServicesToRecover) == 0
}
