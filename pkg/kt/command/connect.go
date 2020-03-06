package command

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// newConnectCommand return new connect command
func newConnectCommand(options *options.DaemonOptions, action Action) cli.Command {

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
			return action.Connect(options)
		},
	}
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(options *options.DaemonOptions) (err error) {
	if util.IsDaemonRunning(options.RuntimeOptions.PidFile) {
		return fmt.Errorf("connect already running %s exit this", options.RuntimeOptions.PidFile)
	}

	ch := SetUpCloseHandler(options)

	pid, err := util.WritePidFile(options.RuntimeOptions.PidFile)
	if err != nil {
		return
	}
	log.Info().Msgf("Connect Start At %d", pid)
	factory := connect.Connect{Options: options}
	clientSet, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return
	}

	if options.ConnectOptions.Dump2Hosts {
		hosts := cluster.LocalHosts(clientSet, options.Namespace)
		util.DumpHosts(hosts)
		options.ConnectOptions.Hosts = hosts
	}

	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	options.RuntimeOptions.Shadow = workload

	labels := map[string]string{
		"kt":           workload,
		"kt-component": "connect",
		"control-by":   "kt",
	}

	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	endPointIP, podName, err := cluster.CreateShadow(
		clientSet,
		workload,
		labels,
		options.Namespace,
		options.Image,
	)

	if err != nil {
		return
	}

	cidrs, err := util.GetCirds(clientSet, options.ConnectOptions.CIDR)
	if err != nil {
		return
	}

	err = factory.StartConnect(podName, endPointIP, cidrs, options.Debug)
	if err != nil {
		return
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return
}
