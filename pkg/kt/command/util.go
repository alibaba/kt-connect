package command

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/signal"
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
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	return
}

// CleanupWorkspace clean workspace
func CleanupWorkspace(cli kt.CliInterface, options *options.DaemonOptions) {
	if !util.IsPidFileExist() {
		log.Info().Msgf("Workspace already cleaned")
		return
	}

	log.Info().Msgf("Cleaning workspace")
	cleanLocalFiles(options)
	removePrivateKey(options)

	if options.RuntimeOptions.Dump2Host {
		log.Debug().Msg("Dropping hosts records ...")
		util.DropHosts()
	}
	if options.ConnectOptions.UseGlobalProxy {
		log.Debug().Msg("Cleaning up global proxy and environment variable ...")
		if options.ConnectOptions.Method == common.ConnectMethodSocks {
			registry.CleanGlobalProxy(&options.RuntimeOptions.ProxyConfig)
		}
		registry.CleanHttpProxyEnvironmentVariable(&options.RuntimeOptions.ProxyConfig)
	}

	if options.ConnectOptions.Method == common.ConnectMethodTun {
		log.Debug().Msg("Removing tun device ...")
		err := exec.RunAndWait(cli.Exec().Tunnel().RemoveDevice(), "del_device")
		if err != nil {
			log.Error().Msgf("Fails to delete tun device")
			return
		}

		if !options.ConnectOptions.DisableDNS {
			err = util.RestoreConfig()
			if err != nil {
				log.Error().Msgf("Restore resolv.conf failed, error: %s", err)
				return
			}
		}
	}

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		log.Error().Msgf("Fails create kubernetes client when clean up workspace")
		return
	}
	ctx := context.Background()
	if len(options.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("Recovering origin deployment %s", options.RuntimeOptions.Origin)
		err := kubernetes.ScaleTo(ctx, options.RuntimeOptions.Origin, options.Namespace, &options.RuntimeOptions.Replicas)
		if err != nil {
			log.Error().
				Str("namespace", options.Namespace).
				Msgf("Scale deployment:%s to %d failed", options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas)
		}
	}

	if len(options.RuntimeOptions.PodName) > 0 {
		log.Info().Msgf("Delete pod: %s", options.RuntimeOptions.PodName)
		err = kubernetes.DeletePod(ctx, options.RuntimeOptions.PodName, options.Namespace)
	}

	cleanDeploymentAndConfigMap(ctx, options, kubernetes)
	cleanService(ctx, options, kubernetes)
}

func cleanLocalFiles(options *options.DaemonOptions) {
	pidFile := fmt.Sprintf("%s/%s-%d.pid", util.KtHome, options.RuntimeOptions.Component, os.Getpid())
	if _, err := os.Stat(pidFile); err == nil {
		log.Info().Msgf("Removing pid %s", pidFile)
		if err = os.Remove(pidFile); err != nil {
			log.Error().Err(err).
				Msgf("Stop process:%s failed", pidFile)
		}
	}

	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		log.Info().Msg("Removing .jvmrc")
		if err := os.Remove(jvmrcFilePath); err != nil {
			log.Error().Err(err).Msg("Delete .jvmrc failed")
		}
	}
}

func cleanService(ctx context.Context, options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	if options.RuntimeOptions.Service != "" {
		log.Info().Msgf("Cleaning service %s", options.RuntimeOptions.Service)
		err := kubernetes.RemoveService(ctx, options.RuntimeOptions.Service, options.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Delete service %s failed", options.RuntimeOptions.Service)
		}
	}
}

func cleanDeploymentAndConfigMap(ctx context.Context, options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	shouldDelWithShared := false
	var err error
	if options.RuntimeOptions.Shadow != "" {
		if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
			shouldDelWithShared, err = decreaseRefOrRemoveTheShadow(ctx, kubernetes, options)
			if err != nil {
				log.Error().Err(err).Msgf("Delete shared deployment %s failed", options.RuntimeOptions.Shadow)
			}
		} else {
			log.Info().Msgf("Cleaning shadow %s", options.RuntimeOptions.Shadow)
			err = kubernetes.RemoveDeployment(ctx, options.RuntimeOptions.Shadow, options.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("Delete deployment %s failed", options.RuntimeOptions.Shadow)
			}
		}
	}

	if options.RuntimeOptions.SSHCM != "" && options.ConnectOptions != nil {
		if shouldDelWithShared || !options.ConnectOptions.ShareShadow {
			log.Info().Msgf("Cleaning config map %s", options.RuntimeOptions.SSHCM)
			err = kubernetes.RemoveConfigMap(ctx, options.RuntimeOptions.SSHCM, options.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("Delete configmap %s failed", options.RuntimeOptions.SSHCM)
			}
		}
	}
}

// decreaseRefOrRemoveTheShadow
func decreaseRefOrRemoveTheShadow(ctx context.Context, kubernetes cluster.KubernetesInterface, options *options.DaemonOptions) (bool, error) {
	return kubernetes.DecreaseRef(ctx, options.Namespace, options.RuntimeOptions.Shadow)
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
		log.Error().Err(err).Msgf("Can't delete %s", file)
	}
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

	config, err := clientcmd.BuildConfigFromFlags("", options.KubeConfig)
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
