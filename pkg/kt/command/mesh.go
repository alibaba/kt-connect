package command

import (
	"errors"

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

			swap := c.Args().First()
			expose := options.MeshOptions.Expose

			if len(swap) == 0 {
				return errors.New("mesh target is required")
			}

			if len(expose) == 0 {
				return errors.New("-expose is required")
			}

			return action.Mesh(c.Args().First(), options)
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, options *options.DaemonOptions) error {
	checkConnectRunning(options.RuntimeOptions.PidFile)

	ch := SetUpCloseHandler(options)

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
