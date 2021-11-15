package general

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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// CleanupWorkspace clean workspace
func CleanupWorkspace(cli kt.CliInterface, opts *options.DaemonOptions) {
	if !util.IsPidFileExist() {
		log.Info().Msgf("Workspace already cleaned")
		return
	}

	log.Info().Msgf("Cleaning workspace")
	cleanLocalFiles(opts)
	removePrivateKey(opts)
	if opts.RuntimeOptions.Component == common.ComponentConnect {
		recoverGlobalHostsAndProxy(opts)
		removeTunDevice(cli, opts)
	}

	ctx := context.Background()
	k8s, err := cli.Kubernetes()
	if err != nil {
		log.Error().Err(err).Msgf("Fails create kubernetes client when clean up workspace")
		return
	}
	if opts.RuntimeOptions.Component == common.ComponentExchange {
		recoverExchangedTarget(ctx, opts, k8s)
	} else if opts.RuntimeOptions.Component == common.ComponentMesh {
		recoverAutoMeshRoute(ctx, opts, k8s)
	} else if opts.RuntimeOptions.Component == common.ComponentProvide {
		cleanService(ctx, opts, k8s)
	}
	cleanShadowPodAndConfigMap(ctx, opts, k8s)
}

func removeTunDevice(cli kt.CliInterface, options *options.DaemonOptions) {
	if options.ConnectOptions.Method == common.ConnectMethodTun {
		log.Debug().Msg("Removing tun device ...")
		err := exec.RunAndWait(cli.Exec().Tunnel().RemoveDevice(), "del_device")
		if err != nil {
			log.Error().Err(err).Msgf("Fails to delete tun device")
		}

		if !options.ConnectOptions.DisableDNS {
			err = util.RestoreConfig()
			if err != nil {
				log.Error().Err(err).Msgf("Restore resolv.conf failed")
			}
		}
	}
}

func recoverGlobalHostsAndProxy(options *options.DaemonOptions) {
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
}

func cleanLocalFiles(options *options.DaemonOptions) {
	pidFile := fmt.Sprintf("%s/%s-%d.pid", util.KtHome, options.RuntimeOptions.Component, os.Getpid())
	if _, err := os.Stat(pidFile); err == nil {
		log.Info().Msgf("Removing pid %s", pidFile)
		if err = os.Remove(pidFile); err != nil {
			log.Error().Err(err).Msgf("Stop process %s failed", pidFile)
		}
	}

	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		log.Info().Msg("Removing .jvmrc")
		if err := os.Remove(jvmrcFilePath); err != nil {
			log.Error().Err(err).Msgf("Delete .jvmrc failed")
		}
	}
}

func recoverExchangedTarget(ctx context.Context, opts *options.DaemonOptions, k8s cluster.KubernetesInterface) {
	if opts.ExchangeOptions.Method == common.ExchangeMethodScale && len(opts.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("Recovering origin deployment %s", opts.RuntimeOptions.Origin)
		err := k8s.ScaleTo(ctx, opts.RuntimeOptions.Origin, opts.Namespace, &opts.RuntimeOptions.Replicas)
		if err != nil {
			log.Error().Err(err).Msgf("Scale deployment %s to %d failed",
				opts.RuntimeOptions.Origin, opts.RuntimeOptions.Replicas)
		}
		// wait for scale complete
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, syscall.SIGINT)
		go func() {
			waitDeploymentRecoverComplete(ctx, opts, k8s)
			ch <- syscall.SIGSTOP
		}()
		_ = <-ch
	}
}

func recoverAutoMeshRoute(ctx context.Context, opts *options.DaemonOptions, k8s cluster.KubernetesInterface) {

}

func waitDeploymentRecoverComplete(ctx context.Context, opts *options.DaemonOptions, k8s cluster.KubernetesInterface) {
	ok := false
	counts := opts.ExchangeOptions.RecoverWaitTime / 5
	for i := 0; i < counts; i++ {
		deployment, err := k8s.GetDeployment(ctx, opts.RuntimeOptions.Origin, opts.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Cannot fetch original deployment %s", opts.RuntimeOptions.Origin)
			break
		} else if deployment.Status.ReadyReplicas == opts.RuntimeOptions.Replicas {
			ok = true
			break
		} else {
			log.Info().Msgf("Wait for deployment %s recover ...", opts.RuntimeOptions.Origin)
			time.Sleep(5 * time.Second)
		}
	}
	if !ok {
		log.Warn().Msgf("Deployment %s recover timeout", opts.RuntimeOptions.Origin)
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

func cleanShadowPodAndConfigMap(ctx context.Context, options *options.DaemonOptions, k8s cluster.KubernetesInterface) {
	shouldDelWithShared := false
	var err error
	if options.RuntimeOptions.Shadow != "" {
		if options.ConnectOptions != nil && options.ConnectOptions.ShareShadow {
			shouldDelWithShared, err = decreaseRefOrRemoveTheShadow(ctx, k8s, options)
			if err != nil {
				log.Error().Err(err).Msgf("Delete shared shadow pod %s failed", options.RuntimeOptions.Shadow)
			}
		} else {
			if options.ExchangeOptions != nil && options.ExchangeOptions.Method == common.ExchangeMethodEphemeral {
				for _, shadow := range strings.Split(options.RuntimeOptions.Shadow, ",") {
					log.Info().Msgf("Removing ephemeral container of pod %s", shadow)
					err = k8s.RemoveEphemeralContainer(ctx, common.KtExchangeContainer, shadow, options.Namespace)
					if err != nil {
						log.Error().Err(err).Msgf("Remove ephemeral container of pod %s failed", shadow)
					}
				}
			} else {
				for _, shadow := range strings.Split(options.RuntimeOptions.Shadow, ",") {
					log.Info().Msgf("Cleaning shadow pod %s", shadow)
					err = k8s.RemovePod(ctx, shadow, options.Namespace)
					if err != nil {
						log.Error().Err(err).Msgf("Delete shadow pod %s failed", shadow)
					}
				}
			}
		}
	}

	if options.RuntimeOptions.SSHCM != "" && options.ConnectOptions != nil && (shouldDelWithShared || !options.ConnectOptions.ShareShadow) {
		for _, sshcm := range strings.Split(options.RuntimeOptions.SSHCM, ",") {
			log.Info().Msgf("Cleaning configmap %s", sshcm)
			err = k8s.RemoveConfigMap(ctx, sshcm, options.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("Delete configmap %s failed", sshcm)
			}
		}
	}
}

// decreaseRefOrRemoveTheShadow
func decreaseRefOrRemoveTheShadow(ctx context.Context, kubernetes cluster.KubernetesInterface, options *options.DaemonOptions) (bool, error) {
	return kubernetes.DecreaseRef(ctx, options.RuntimeOptions.Shadow, options.Namespace)
}

// removePrivateKey remove the private key of ssh
func removePrivateKey(options *options.DaemonOptions) {
	if options.RuntimeOptions.SSHCM == "" {
		return
	}
	for _, sshcm := range strings.Split(options.RuntimeOptions.SSHCM, ",") {
		splits := strings.Split(sshcm, "-")
		component, version := splits[1], splits[len(splits)-1]
		file := util.PrivateKeyPath(component, version)
		if err := os.Remove(file); os.IsNotExist(err) {
			log.Error().Msgf("Key file %s not exist", file)
		}
	}
}
