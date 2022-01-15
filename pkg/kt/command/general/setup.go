package general

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

// SetupProcess write pid file and print setup message
func SetupProcess(cli kt.CliInterface, options *options.DaemonOptions, componentName string) (chan os.Signal, error) {
	options.RuntimeOptions.Component = componentName
	err := util.WritePidFile(componentName)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("KtConnect %s start at %d (%s)", options.RuntimeOptions.Version, os.Getpid(), runtime.GOOS)
	ch := setupCloseHandler(cli, options)
	return ch, nil
}

// setupWaitingChannel registry waiting channel
func setupWaitingChannel() (ch chan os.Signal) {
	ch = make(chan os.Signal)
	// see https://en.wikipedia.org/wiki/Signal_(IPC)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	return
}

// SetupCloseHandler registry close handler
func setupCloseHandler(cli kt.CliInterface, options *options.DaemonOptions) (ch chan os.Signal) {
	ch = setupWaitingChannel()
	go func() {
		<-ch
		log.Info().Msgf("Process is gonna close")
		CleanupWorkspace(cli, options)
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

// CombineKubeOpts set default options of kubectl if not assign
func CombineKubeOpts(options *options.DaemonOptions) error {
	config, err := clientcmd.LoadFromFile(options.KubeConfig)
	if err != nil {
		return err
	}
	if len(options.KubeContext) > 0 {
		found := false
		for name, _ := range config.Contexts {
			if name == options.KubeContext {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context '%s' not exist, check your kubeconfig file please", options.KubeContext)
		}
		config.CurrentContext = options.KubeContext
	}
	if len(options.Namespace) == 0 {
		if len(config.Contexts[config.CurrentContext].Namespace) > 0 {
			options.Namespace = config.Contexts[config.CurrentContext].Namespace
		} else {
			options.Namespace = common.DefaultNamespace
		}
	}
	kubeconfigGetter := func() clientcmd.KubeconfigGetter {
		return func() (*clientcmdapi.Config, error) {
			return config, nil
		}
	}
	restConfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", kubeconfigGetter())
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	options.RuntimeOptions.Clientset = clientSet
	options.RuntimeOptions.RestConfig = restConfig

	return nil
}
