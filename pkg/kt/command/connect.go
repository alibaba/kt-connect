package command

import (
	"fmt"
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
func (action *Action) Connect(cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	if util.IsDaemonRunning(options.RuntimeOptions.PidFile) {
		return fmt.Errorf("connect already running %s exit this", options.RuntimeOptions.PidFile)
	}
	ch := SetUpCloseHandler(cli, options, "connect")
	if err = connectToCluster(cli, options); err != nil {
		return
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-util.Interrupt()
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return
}

func connectToCluster(cli kt.CliInterface, options *options.DaemonOptions) (err error) {

	pid, err := util.WritePidFile(options.RuntimeOptions.PidFile)
	if err != nil {
		return
	}
	log.Info().Msgf("Connect Start At %d", pid)

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return
	}

	if options.ConnectOptions.Dump2Hosts {
		setupDump2Host(options, kubernetes)
	}

	endPointIP, podName, credential, err := getOrCreateShadow(options, err, kubernetes)
	if err != nil {
		return
	}

	cidrs, err := kubernetes.ClusterCrids(options.Namespace, options.ConnectOptions)
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

	endPointIP, podName, sshcm, credential, err :=
		kubernetes.GetOrCreateShadow(workload, options.Namespace, options.Image, labels(workload, options), envs(options),
			options.Debug, options.ConnectOptions.ShareShadow)
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
	if options.ConnectOptions.Dump2HostsNamespaces != nil {
		for _, namespace := range options.ConnectOptions.Dump2HostsNamespaces {
			if namespace == options.Namespace {
				continue
			}
			log.Debug().Msgf("Search service in %s namespace...", namespace)
			singleHosts := kubernetes.ServiceHosts(namespace)
			for k, v := range singleHosts {
				if v == "" || v == "None" {
					continue
				}
				log.Info().Msgf("Service found: %s.%s %s", k, namespace, v)
				hosts[k+"."+namespace] = v
			}
		}
	}
	util.DumpHosts(hosts)
	options.ConnectOptions.Hosts = hosts
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
