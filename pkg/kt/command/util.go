package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// NewCommands return new Connect Action
func NewCommands(kt kt.CliInterface, action ActionInterface, options *options.DaemonOptions) []cli.Command {
	return []cli.Command{
		newConnectCommand(kt, options, action),
		newExchangeCommand(kt, options, action),
		newMeshCommand(kt, options, action),
		newProvideCommand(kt, options, action),
		newCleanCommand(kt, options, action),
		newDashboardCommand(kt, options, action),
		newCheckCommand(kt, options, action),
	}
}

// SetUpWaitingChannel registry waiting channel
func SetUpWaitingChannel() (ch chan os.Signal) {
	ch = make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return
}

// SetUpCloseHandler registry close handler
func SetUpCloseHandler(cli kt.CliInterface, options *options.DaemonOptions, action string) (ch chan os.Signal) {
	ch = make(chan os.Signal)
	// see https://en.wikipedia.org/wiki/Signal_(IPC)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-ch
		log.Info().Msgf("- Terminal And Clean Workspace")
		CleanupWorkspace(cli, options)
		log.Info().Msgf("- Successful Clean Up Workspace")
		os.Exit(0)
	}()
	return
}

// CleanupWorkspace clean workspace
func CleanupWorkspace(cli kt.CliInterface, options *options.DaemonOptions) {
	log.Info().Msgf("- start Clean Workspace")
	cleanLocalFiles(options)
	removePrivateKey(options)

	if options.RuntimeOptions.Dump2Host {
		util.DropHosts()
	}
	if options.ConnectOptions.Method == common.ConnectMethodSocks {
		registry.CleanGlobalProxy(&options.RuntimeOptions.ProxyConfig)
		registry.CleanHttpProxyEnvironmentVariable(&options.RuntimeOptions.ProxyConfig)
	}

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		log.Error().Msgf("fails create kubernetes client when clean up workspace")
		return
	}

	if len(options.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("- Recover Origin App %s", options.RuntimeOptions.Origin)
		err := kubernetes.ScaleTo(options.RuntimeOptions.Origin, options.Namespace, &options.RuntimeOptions.Replicas)
		if err != nil {
			log.Error().
				Str("namespace", options.Namespace).
				Msgf("scale deployment:%s to %d failed", options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas)
		}
	}

	cleanDeploymentAndConfigMap(options, kubernetes)
	cleanService(options, kubernetes)
}

func cleanLocalFiles(options *options.DaemonOptions) {
	if _, err := os.Stat(options.RuntimeOptions.PidFile); err == nil {
		log.Info().Msgf("- Remove pid %s", options.RuntimeOptions.PidFile)
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
}

func cleanService(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	if options.RuntimeOptions.Service != "" {
		log.Info().Msgf("- Cleanup service %s", options.RuntimeOptions.Service)
		err := kubernetes.RemoveService(options.RuntimeOptions.Service, options.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("delete service %s failed", options.RuntimeOptions.Service)
		}
	}
}

func cleanDeploymentAndConfigMap(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	shouldDelWithShared := false
	var err error
	if options.RuntimeOptions.Shadow != "" {
		if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
			shouldDelWithShared, err = decreaseRefOrRemoveTheShadow(kubernetes, options)
			if err != nil {
				log.Error().Err(err).Msgf("delete shared deployment %s failed", options.RuntimeOptions.Shadow)
			}
		} else {
			log.Info().Msgf("- clean shadow %s", options.RuntimeOptions.Shadow)
			err = kubernetes.RemoveDeployment(options.RuntimeOptions.Shadow, options.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("delete deployment %s failed", options.RuntimeOptions.Shadow)
			}
		}
	}

	if options.RuntimeOptions.SSHCM != "" && options.ConnectOptions != nil {
		if shouldDelWithShared || !options.ConnectOptions.ShareShadow {
			log.Info().Msgf("- clean sshcm %s", options.RuntimeOptions.SSHCM)
			err = kubernetes.RemoveConfigMap(options.RuntimeOptions.SSHCM, options.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("delete configmap %s failed", options.RuntimeOptions.SSHCM)
			}
		}
	}
}

// decreaseRefOrRemoveTheShadow
func decreaseRefOrRemoveTheShadow(kubernetes cluster.KubernetesInterface, options *options.DaemonOptions) (bool, error) {
	return kubernetes.DecreaseRef(options.Namespace, options.RuntimeOptions.Shadow)
}

// removePrivateKey remove the private key of ssh
func removePrivateKey(options *options.DaemonOptions) {
	if options.RuntimeOptions.SSHCM == "" {
		return
	}
	splits := strings.Split(options.RuntimeOptions.SSHCM, "-")
	component, version := splits[1], splits[len(splits)-1]
	file := util.PrivateKeyPath(component, version)
	if err := os.Remove(file); os.IsNotExist(err) {
		log.Error().Err(err).Msgf("can't delete %s", file)
	}
}

// validateKubeOpts support like '-n default | --kubeconfig=/path/to/kubeconfig'
func validateKubeOpts(opts []string) error {
	errMsg := "kubectl option %s invalid, check it by 'kubectl options'"
	for _, opt := range opts {
		// validate like '--kubeconfig=/path/to/kube/config'
		if strings.Contains(opt, "=") && len(strings.Fields(opt)) != 1 {
			return fmt.Errorf(errMsg, opt)
		}
		// validate like '-n default'
		if strings.Contains(opt, " ") && len(strings.Fields(opt)) != 2 {
			return fmt.Errorf(errMsg, opt)
		}
	}
	return nil
}

// combineKubeOpts set default options of kubectl if not assign
func combineKubeOpts(options *options.DaemonOptions) error {
	if err := validateKubeOpts(options.KubeOptions); err != nil {
		return err
	}

	var configured, namespaced bool
	for _, opt := range options.KubeOptions {
		strs := strings.Fields(opt)
		if len(strs) == 1 {
			strs = strings.Split(opt, "=")
		}
		switch strs[0] {
		case "-n", "--namespace":
			options.Namespace = strs[1]
			namespaced = true
		case "--kubeconfig":
			options.KubeConfig = strs[1]
			configured = true
		}
	}

	if !configured {
		options.KubeOptions = append(options.KubeOptions, fmt.Sprintf("--kubeconfig=%s", options.KubeConfig))
	}

	if !namespaced {
		options.KubeOptions = append(options.KubeOptions, fmt.Sprintf("--namespace=%s", options.Namespace))
	}
	return nil
}
