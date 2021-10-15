package clean

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
	"strings"
)

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

	k8s, err := cli.Kubernetes()
	if err != nil {
		log.Error().Msgf("Fails create kubernetes client when clean up workspace")
		return
	}
	ctx := context.Background()
	if len(options.RuntimeOptions.Origin) > 0 {
		log.Info().Msgf("Recovering origin deployment %s", options.RuntimeOptions.Origin)
		err := k8s.ScaleTo(ctx, options.RuntimeOptions.Origin, options.Namespace, &options.RuntimeOptions.Replicas)
		if err != nil {
			log.Error().
				Str("namespace", options.Namespace).
				Msgf("Scale deployment:%s to %d failed", options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas)
		}
	}

	cleanShadowPodAndConfigMap(ctx, options, k8s)
	cleanService(ctx, options, k8s)
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
	return kubernetes.DecreaseRef(ctx, options.Namespace, options.RuntimeOptions.Shadow)
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
			log.Error().Err(err).Msgf("Can't delete %s", file)
		}
	}
}
