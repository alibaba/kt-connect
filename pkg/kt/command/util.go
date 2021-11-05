package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
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
	}
}

// setupProcess write pid file and print setup message
func setupProcess(cli kt.CliInterface, options *options.DaemonOptions, componentName string) (chan os.Signal, error) {
	options.RuntimeOptions.Component = componentName
	err := util.WritePidFile(componentName)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("KtConnect %s start at %d (%s)", options.Version, os.Getpid(), runtime.GOOS)
	ch := SetupCloseHandler(cli, options, common.ComponentProvide)
	return ch, nil
}

// setupWaitingChannel registry waiting channel
func setupWaitingChannel() (ch chan os.Signal) {
	ch = make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return
}

// SetupCloseHandler registry close handler
func SetupCloseHandler(cli kt.CliInterface, options *options.DaemonOptions, action string) (ch chan os.Signal) {
	ch = make(chan os.Signal)
	// see https://en.wikipedia.org/wiki/Signal_(IPC)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-ch
		clean.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	return
}

// validateKubeOpts support like '-n default | --kubeconfig=/path/to/kubeconfig'
func validateKubeOpts(opts []string) error {
	errMsg := "Kubectl option %s invalid, check it by 'kubectl options'"
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

	kubeconfigGetter := func(configFile string) clientcmd.KubeconfigGetter {
		return func() (*clientcmdapi.Config, error) {
			config, err := clientcmd.LoadFromFile(configFile)
			if err != nil {
				return nil, err
			}
			return config, nil
		}
	}
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", kubeconfigGetter(options.KubeConfig))
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	options.RuntimeOptions.Clientset = clientset
	options.RuntimeOptions.RestConfig = config

	return nil
}
