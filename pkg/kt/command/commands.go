package command

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// NewCliAuthor return cli author
func NewCliAuthor() []cli.Author {
	return []cli.Author{
		cli.Author{
			Name: "rdc incubator",
		},
	}
}

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

// newMeshCommand return new mesh command
func newMeshCommand(options *options.DaemonOptions) cli.Command {
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
			action := Action{}
			action.Mesh(c.Args().First(), options)
			return nil
		},
	}
}

// NewCheckCommand return new check command
func NewCheckCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "check",
		Usage: "check local dependency for ktctl",
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action := Action{}
			return action.Check(options)
		},
	}
}

// NewCommands return new Connect Command
func NewCommands(options *options.DaemonOptions) []cli.Command {
	return []cli.Command{
		NewCheckCommand(options),
		newConnectCommand(options),
		newExchangeCommand(options),
		newMeshCommand(options),
	}
}

// AppFlags return app flags
func AppFlags(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Value:       "default",
			Destination: &options.Namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Value:       filepath.Join(options.RuntimeOptions.UserHome, ".kube", "config"),
			Destination: &options.KubeConfig,
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

// SetUpCloseHandler registry close handeler
func SetUpCloseHandler(options *options.DaemonOptions) (ch chan os.Signal) {
	ch = make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-ch
		log.Info().Msgf("- Terminal And Clean Workspace\n")
		CleanupWorkspace(options)
		log.Info().Msgf("- Successful Clean Up Workspace\n")
		os.Exit(0)
	}()
	return
}

// CleanupWorkspace clean workspace
func CleanupWorkspace(options *options.DaemonOptions) {
	log.Info().Msgf("- Start Clean Workspace\n")
	if _, err := os.Stat(options.RuntimeOptions.PidFile); err == nil {
		log.Info().Msgf("- Remove pid %s", options.RuntimeOptions.PidFile)
		os.Remove(options.RuntimeOptions.PidFile)
	}

	if _, err := os.Stat(".jvmrc"); err == nil {
		log.Info().Msgf("- Remove .jvmrc %s", options.RuntimeOptions.PidFile)
		os.Remove(".jvmrc")
	}
	util.DropHosts(options.ConnectOptions.Hosts)
	client, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		log.Error().Msgf("Fails create kubernetes client when clean up workspace")
		return
	}

	// scale origin app to replicas
	if len(options.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("- Recover Origin App %s", options.RuntimeOptions.Origin)
		cluster.ScaleTo(
			client,
			options.Namespace,
			options.RuntimeOptions.Origin,
			options.RuntimeOptions.Replicas,
		)
	}

	if len(options.RuntimeOptions.Shadow) > 0 {
		log.Info().Msgf("- Start Clean Shadow %s", options.RuntimeOptions.Shadow)
		cluster.Remove(client, options.Namespace, options.RuntimeOptions.Shadow)
		log.Info().Msgf("- Successful Clean Shadow %s", options.RuntimeOptions.Shadow)
	}
}
