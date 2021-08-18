package command

import (
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
)

// newDashboardCommand dashboard command
func newDashboardCommand(ktCli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) cli.Command {
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
					if err := combineKubeOpts(options); err != nil {
						return err
					}
					return action.ApplyDashboard(ktCli, options)
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
					if err := combineKubeOpts(options); err != nil {
						return err
					}
					return action.OpenDashboard(ktCli, options)
				},
			},
		},
	}
}

// ApplyDashboard ...
func (action *Action) ApplyDashboard(cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	command := cli.Exec().Kubectl().ApplyDashboardToCluster()
	log.Info().Msg("Install/Upgrade Dashboard to cluster")
	err = exec.RunAndWait(command, "apply kt dashboard")
	if err != nil {
		log.Error().Msg("Fail to apply dashboard, please check the log")
		return
	}
	return
}

// OpenDashboard ...
func (action *Action) OpenDashboard(ktCli kt.CliInterface, options *options.DaemonOptions) (err error) {
	ch := SetUpWaitingChannel()
	command := ktCli.Exec().Kubectl().PortForwardDashboardToLocal(options.DashboardOptions.Port)
	err = exec.BackgroundRun(command, "forward dashboard to localhost")
	if err != nil {
		return
	}
	err = open.Run("http://127.0.0.1:" + options.DashboardOptions.Port)

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return
}
