package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/action"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
	"path/filepath"
)

// NewCliAuthor return cli author
func NewCliAuthor() []cli.Author {
	return []cli.Author{
		cli.Author{
			Name: "rdc incubator",
		},
	}
}

func newConnectCommand(options *DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "method",
				Value:       "vpn",
				Usage:       "Connect method 'vpn' or 'socks5'",
				Destination: &options.ConnectOptions.Method,
			},
			cli.IntFlag{
				Name:        "proxy",
				Value:       2223,
				Usage:       "when should method socks5, you can choice which port to proxy, default 2223",
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
		},
		Action: func(c *cli.Context) error {
			action := action.Action{
				Kubeconfig: options.Kubeconfig,
				Namespace:  options.Namespace,
				Debug:      options.Debug,
				Image:      options.Image,
				PidFile:    options.RuntimeOptions.PidFile,
				UserHome:   options.RuntimeOptions.UserHome,
				Labels:     options.Labels,
			}
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action.Connect(options.ConnectOptions.SSHPort, options.ConnectOptions.Method, options.ConnectOptions.Socke5Proxy, options.ConnectOptions.DisableDNS, options.ConnectOptions.CIDR)
			return nil
		},
	}
}

func newExchangeCommand(options *DaemonOptions) cli.Command {
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
			action := action.Action{
				Kubeconfig: options.Kubeconfig,
				Namespace:  options.Namespace,
				Debug:      options.Debug,
				Image:      options.Image,
				Labels:     options.Labels,
				PidFile:    options.RuntimeOptions.PidFile,
				UserHome:   options.RuntimeOptions.UserHome,
			}
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action.Exchange(c.Args().First(), options.ExchangeOptions.Expose, options.RuntimeOptions.UserHome, options.RuntimeOptions.PidFile)
			return nil
		},
	}
}

func newMeshCommand(options *DaemonOptions) cli.Command {
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
			action := action.Action{
				Kubeconfig: options.Kubeconfig,
				Namespace:  options.Namespace,
				Debug:      options.Debug,
				Image:      options.Image,
				Labels:     options.Labels,
				PidFile:    options.RuntimeOptions.PidFile,
				UserHome:   options.RuntimeOptions.UserHome,
			}
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action.Mesh(c.Args().First(), options.MeshOptions.Expose, options.RuntimeOptions.UserHome, options.RuntimeOptions.PidFile)
			return nil
		},
	}
}

// NewCommands return new Connect Command
func NewCommands(options *DaemonOptions) []cli.Command {
	return []cli.Command {
		newConnectCommand(options),
		newExchangeCommand(options),
		newMeshCommand(options),
	}
}

// AppFlags return app flags
func AppFlags(options *DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Value:       "default",
			Destination: &options.Namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Value:       filepath.Join(options.RuntimeOptions.UserHome, ".kube", "config"),
			Destination: &options.Kubeconfig,
		},
		cli.StringFlag{
			Name:        "image,i",
			Usage:       "Custom proxy image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:stable",
			Destination: &options.Image,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "debug mode",
			Destination: &options.Debug,
		},
		cli.StringFlag{
			Name:        "label,l",
			Usage:       "Extra labels on proxy pod e.g. 'label1=val1,label2=val2'",
			Destination: &options.Labels,
		},
	}
}
