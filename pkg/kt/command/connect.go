package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
)

// newConnectCommand return new connect command
func newConnectCommand(options *options.DaemonOptions) cli.Command {

	methodDefaultValue := "vpn"
	methodDefaultUsage := "Connect method 'vpn' or 'socks5'"
	if util.IsWindows() {
		methodDefaultValue = "socks5"
		methodDefaultUsage = "windows only support socks5"
	}

	return cli.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "method",
				Value:       methodDefaultValue,
				Usage:       methodDefaultUsage,
				Destination: &options.ConnectOptions.Method,
			},
			cli.IntFlag{
				Name:        "proxy",
				Value:       2223,
				Usage:       "when should method socks5, you can choice which port to proxy",
				Destination: &options.ConnectOptions.Socke5Proxy,
			},
			cli.IntFlag{
				Name:        "port",
				Value:       2222,
				Usage:       "Local SSH Proxy port",
				Destination: &options.ConnectOptions.SSHPort,
			},
			cli.BoolFlag{
				Name:        "disableDNS",
				Usage:       "Disable Cluster DNS",
				Destination: &options.ConnectOptions.DisableDNS,
			},
			cli.StringFlag{
				Name:        "cidr",
				Usage:       "Custom CIDR eq '172.2.0.0/16'",
				Destination: &options.ConnectOptions.CIDR,
			},
			cli.BoolFlag{
				Name:        "dump2hosts",
				Usage:       "Auto write service to local hosts file",
				Destination: &options.ConnectOptions.Dump2Hosts,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action := Action{}
			return action.Connect(options)
		},
	}
}
