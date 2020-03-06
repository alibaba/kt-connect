package command

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// newMeshCommand return new mesh command
func newMeshCommand(options *options.DaemonOptions, action ActionInterface) cli.Command {
	return cli.Command{
		Name:  "mesh",
		Usage: "mesh kubernetes deployment to local",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "expose",
				Usage:       "expose port",
				Destination: &options.MeshOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action.Mesh(c.Args().First(), options)
			return nil
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, options *options.DaemonOptions) error {
	checkConnectRunning(options.RuntimeOptions.PidFile)

	ch := SetUpCloseHandler(options)

	expose := options.MeshOptions.Expose

	if swap == "" || expose == "" {
		return fmt.Errorf("-expose is required")
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return err
	}

	factory := connect.Connect{}
	_, err = factory.Mesh(swap, options, clientset, util.String2Map(options.Labels))

	if err != nil {
		return err
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}
