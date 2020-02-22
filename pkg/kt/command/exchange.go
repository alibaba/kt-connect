package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
)

// newExchangeCommand return new exchange command
func newExchangeCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "expose",
				Usage:       "expose port",
				Destination: &options.ExchangeOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			action := Action{}
			return action.Exchange(c.Args().First(), options)
		},
	}
}
