package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// NewCleanCommand return new connect command
func NewCleanCommand(action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "clean",
		Usage: "delete unavailing resources created by kt from kubernetes cluster",
		UsageText: "ktctl clean [command options]",
		Flags: general.CleanActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}
			return action.Clean()
		},
	}
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
