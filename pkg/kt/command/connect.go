package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// newConnectCommand return new connect command
func newConnectCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: ConnectActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			return action.Connect(cli, options)
		},
	}
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(cli kt.CliInterface, options *options.DaemonOptions) error {
	if util.IsDaemonRunning(common.ComponentConnect) {
		return fmt.Errorf("another connect process already running, exiting")
	}

	options.RuntimeOptions.Component = common.ComponentConnect
	err := util.WritePidFile(common.ComponentConnect)
	if err != nil {
		return err
	}
	log.Info().Msgf("KtConnect start at %d", os.Getpid())

	ch := SetUpCloseHandler(cli, options, common.ComponentConnect)
	if err = connectToCluster(cli, options); err != nil {
		return err
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		log.Error().Msgf("Command interrupted: %s", <-util.Interrupt())
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal signal is %s", s)
	return nil
}

func connectToCluster(cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return
	}

	if util.IsWindows() || len(options.ConnectOptions.Dump2HostsNamespaces) > 0 {
		setupDump2Host(options, kubernetes)
	}
	if options.ConnectOptions.Method == common.ConnectMethodSocks {
		err = registry.SetGlobalProxy(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
		if err != nil {
			log.Error().Msgf("Failed to setup global connect proxy: %s", err.Error())
		}
		err = registry.SetHttpProxyEnvironmentVariable(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
		if err != nil {
			log.Error().Msgf("Failed to setup global http proxy: %s", err.Error())
		}
	}

	endPointIP, podName, credential, err := getOrCreateShadow(options, err, kubernetes)
	if err != nil {
		return
	}

	cidrs, err := kubernetes.ClusterCidrs(options.Namespace, options.ConnectOptions)
	if err != nil {
		return
	}

	return cli.Shadow().Outbound(podName, endPointIP, credential, cidrs, cli.Exec())
}

func getOrCreateShadow(options *options.DaemonOptions, err error, kubernetes cluster.KubernetesInterface) (string, string, *util.SSHCredential, error) {
	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.ShareShadow {
		workload = fmt.Sprintf("kt-connect-daemon-connect-shared")
	}

	annotations := make(map[string]string)
	endPointIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(workload, options, labels(workload, options), annotations, envs(options))
	if err != nil {
		return "", "", nil, err
	}

	// record shadow name will clean up terminal
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	return endPointIP, podName, credential, nil
}

func setupDump2Host(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	hosts := kubernetes.ServiceHosts(options.Namespace)
	for k, v := range hosts {
		log.Info().Msgf("Service found: %s %s", k, v)
	}
	if len(options.ConnectOptions.Dump2HostsNamespaces) > 0 {
		for _, namespace := range options.ConnectOptions.Dump2HostsNamespaces {
			if namespace == options.Namespace {
				continue
			}
			log.Debug().Msgf("Search service in %s namespace...", namespace)
			singleHosts := kubernetes.ServiceHosts(namespace)
			for svc, ip := range singleHosts {
				if ip == "" || ip == "None" {
					continue
				}
				log.Info().Msgf("Service found: %s.%s %s", svc, namespace, ip)
				hosts[svc+"."+namespace] = ip
			}
		}
	}
	util.DumpHosts(hosts)
	options.RuntimeOptions.Dump2Host = true
}

func envs(options *options.DaemonOptions) map[string]string {
	envs := make(map[string]string)
	if options.ConnectOptions.LocalDomain != "" {
		envs[common.EnvVarLocalDomain] = options.ConnectOptions.LocalDomain
	}
	return envs
}

func labels(workload string, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentConnect,
		common.KTName:      workload,
	}
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}
	splits := strings.Split(workload, "-")
	labels[common.KTVersion] = splits[len(splits)-1]
	return labels
}
