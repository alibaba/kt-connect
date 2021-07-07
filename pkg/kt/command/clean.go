package command

import (
	"container/list"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"k8s.io/api/apps/v1"
	"strconv"
	"time"
)

// newConnectCommand return new connect command
func newCleanCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "clean",
		Usage: "delete unavailing shadow pods from kubernetes cluster",
		Flags: []urfave.Flag{
			urfave.BoolFlag{
				Name:        "dryRun",
				Usage:       "Only print name of deployments to be deleted",
				Destination: &options.CleanOptions.DryRun,
			},
			urfave.Int64Flag{
				Name:        "thresholdInMinus",
				Usage:       "Length of allowed disconnection time before a unavailing shadow pod be deleted",
				Destination: &options.CleanOptions.ThresholdInMinus,
				Value:       30,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			return action.Clean(cli, options)
		},
	}
}

//Clean delete unavailing shadow pods
func (action *Action) Clean(cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, deployments, err := action.getShadowDeployments(cli, options)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Found %d shadow deployments", len(deployments))
	namesOfDeploymentToDelete := list.New()
	for _, deployment := range deployments {
		lastHeartBeat, err := strconv.ParseInt(deployment.ObjectMeta.Annotations[common.KTLastHeartBeat], 10, 64)
		if err == nil && action.isExpired(lastHeartBeat, options) {
			namesOfDeploymentToDelete.PushBack(deployment.Name)
		}
	}
	if namesOfDeploymentToDelete.Len() == 0 {
		log.Info().Msg("No unavailing shadow deployment found (^.^)YYa!!")
		return nil
	}
	if options.CleanOptions.DryRun {
		log.Info().Msgf("Found %d unavailing shadow deployments:", namesOfDeploymentToDelete.Len())
		for name := namesOfDeploymentToDelete.Front(); name != nil; name = name.Next() {
			log.Info().Msgf("> %s", name.Value.(string))
		}
	} else {
		log.Info().Msgf("Deleting %d unavailing shadow deployments", namesOfDeploymentToDelete.Len())
		for name := namesOfDeploymentToDelete.Front(); name != nil; name = name.Next() {
			err := kubernetes.RemoveDeployment(name.Value.(string), options.Namespace)
			if err != nil {
				return err
			}
		}
		log.Info().Msg("Done.")
	}
	return nil
}

func (action *Action) isExpired(lastHeartBeat int64, options *options.DaemonOptions) bool {
	return time.Now().Unix()-lastHeartBeat > options.CleanOptions.ThresholdInMinus*60
}

func (action *Action) getShadowDeployments(cli kt.CliInterface, options *options.DaemonOptions) (
	cluster.KubernetesInterface, []v1.Deployment, error) {

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return nil, nil, err
	}
	deployments, err := kubernetes.GetAllExistingShadowDeployments(options.Namespace)
	if err != nil {
		return nil, nil, err
	}
	return kubernetes, deployments, nil
}
