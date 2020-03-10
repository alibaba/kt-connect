package command

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// NewCommands return new Connect Command
func NewCommands(options *options.DaemonOptions, action ActionInterface) []cli.Command {
	return []cli.Command{
		newRunCommand(options, action),
		newConnectCommand(options, action),
		newExchangeCommand(options, action),
		newMeshCommand(options, action),
		newDashboardCommand(options, action),
		NewCheckCommand(options, action),
	}
}

// SetUpWaitingChannel registry waiting channel
func SetUpWaitingChannel() (ch chan os.Signal) {
	ch = make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return
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
	log.Info().Msgf("- start Clean Workspace\n")
	if _, err := os.Stat(options.RuntimeOptions.PidFile); err == nil {
		log.Info().Msgf("- remove pid %s", options.RuntimeOptions.PidFile)
		if err = os.Remove(options.RuntimeOptions.PidFile); err != nil {
			log.Error().Err(err).
				Msgf("stop process:%s failed", options.RuntimeOptions.PidFile)
		}
	}

	if _, err := os.Stat(".jvmrc"); err == nil {
		log.Info().Msgf("- Remove .jvmrc %s", options.RuntimeOptions.PidFile)
		if err = os.Remove(".jvmrc"); err != nil {
			log.Error().Err(err).Msg("delete .jvmrc failed")
		}
	}

	util.DropHosts(options.ConnectOptions.Hosts)
	client, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		log.Error().Msgf("fails create kubernetes client when clean up workspace")
		return
	}

	// scale origin app to replicas
	if len(options.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("- Recover Origin App %s", options.RuntimeOptions.Origin)
		err = cluster.ScaleTo(
			client,
			options.Namespace,
			options.RuntimeOptions.Origin,
			options.RuntimeOptions.Replicas,
		)
		if err != nil {
			log.Error().
				Str("namespace", options.Namespace).
				Msgf("scale deployment:%s to %d failed", options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas)
		}
	}

	if len(options.RuntimeOptions.Shadow) > 0 {
		log.Info().Msgf("- clean shadow %s", options.RuntimeOptions.Shadow)
		cluster.RemoveShadow(client, options.Namespace, options.RuntimeOptions.Shadow)
	}

	if len(options.RuntimeOptions.SSHCM) > 0 {
		log.Info().Msgf("- clean sshcm %s", options.RuntimeOptions.SSHCM)
		cluster.RemoveSSHCM(client, options.Namespace, options.RuntimeOptions.SSHCM)
	}

	removePrivateKey(options)
	if len(options.RuntimeOptions.Service) > 0 {
		log.Info().Msgf("- cleanup service %s", options.RuntimeOptions.Service)
		err = cluster.RemoveService(options.RuntimeOptions.Service, options.Namespace, client)
		if err != nil {
			log.Error().Err(err).Msg("delete service failed")
		}
	}
}

// checkConnectRunning check connect is running and print help msg
func checkConnectRunning(pidFile string) {
	daemonRunning := util.IsDaemonRunning(pidFile)
	if !daemonRunning {
		log.Info().Msgf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Info().Msgf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}
}

// removePrivateKey remove the private key of ssh
func removePrivateKey(options *options.DaemonOptions) {
	if options.RuntimeOptions.SSHCM == "" {
		return
	}
	splits := strings.Split(options.RuntimeOptions.SSHCM, "-")
	component, version := splits[1], splits[len(splits)-1]
	file := util.PrivateKeyPath(component, version)
	if err := os.Remove(file); !os.IsNotExist(err) {
		log.Error().Err(err).Msgf("can't delete %s", file)
	}
}
