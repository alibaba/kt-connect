package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
)

// newDashboardCommand dashboard command
func newDashboardCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "dashboard",
		Usage: "kt-connect dashboard",
		Subcommands: []cli.Command{
			{
				Name:  "init",
				Usage: "install/update dashboard to cluster",
				Action: func(c *cli.Context) error {
					if options.Debug {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					action := Action{}
					return action.ApplyDashboard()
				},
			},
			{
				Name:  "open",
				Usage: "open dashboard",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "port,p",
						Value:       "8080",
						Usage:       "port-forward kt dashboard to port",
						Destination: &options.DashboardOptions.Port,
					},
				},
				Action: func(c *cli.Context) error {
					if options.Debug {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					action := Action{}
					return action.OpenDashboard(options)
				},
			},
		},
	}
}
