package command

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// NewCommands return new Connect Command
func NewCommands(options *options.DaemonOptions) []cli.Command {
	return []cli.Command{
		newDashboardCommand(options),
		newConnectCommand(options),
		newExchangeCommand(options),
		newMeshCommand(options),
		newCheckCommand(options),
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

// checkConnectRunning check connect is running and print help msg
func checkConnectRunning(pidFile string) {
	daemonRunning := util.IsDaemonRunning(pidFile)
	if !daemonRunning {
		log.Info().Msgf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Info().Msgf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}
}