package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewCleanCommand return new connect command
func NewCleanCommand(action ActionInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "clean",
		Long: "Delete unavailing resources created by kt from kubernetes cluster",
		Short: "ktctl clean [command options]",
		Run: func(cmd *cobra.Command, args []string) {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				log.Error().Msgf("%s", err)
			} else if err2 := action.Clean(); err2 != nil {
				log.Error().Msgf("%s", err2)
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().Int64Var(&opt.Get().CleanOptions.ThresholdInMinus, "thresholdInMinus", cluster.ResourceHeartBeatIntervalMinus * 2 + 1, "Length of allowed disconnection time before a unavailing shadow pod be deleted")
	cmd.Flags().BoolVar(&opt.Get().CleanOptions.DryRun, "dryRun", false, "Only print name of deployments to be deleted")
	return cmd
}

// Clean delete unavailing shadow pods
func (action *Action) Clean() error {
	if resourceToClean, err := clean.CheckClusterResources(); err != nil {
		log.Warn().Err(err).Msgf("Failed to clean up cluster resources")
	} else {
		if isEmpty(resourceToClean) {
			log.Info().Msg("No unavailing kt resource found (^.^)YYa!!")
		} else {
			if opt.Get().CleanOptions.DryRun {
				clean.PrintClusterResourcesToClean(resourceToClean)
			} else {
				clean.TidyClusterResources(resourceToClean)
			}
		}
	}
	if !opt.Get().CleanOptions.DryRun {
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
