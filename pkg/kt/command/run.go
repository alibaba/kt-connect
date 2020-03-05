package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// newRunCommand return new run command
func newRunCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "run",
		Usage: "Create a shadow deployment to redirect request to user local",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:        "port",
				Usage:       "The port that exposes",
				Destination: &options.RunOptions.Port,
			},
			cli.BoolFlag{
				Name:        "expose",
				Usage:       " If true, a public, external service is created",
				Destination: &options.RunOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action := Action{}
			action.Run(c.Args().First(), options)
			return nil
		},
	}
}

// Run create a new service in cluster
func (action *Action) Run(service string, options *options.DaemonOptions) error {
	log.Info().Msgf("run service %s port %d expose %b", service, options.RunOptions.Port, options.RunOptions.Expose)
	return nil
}
