package general

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/dns"
	"github.com/alibaba/kt-connect/pkg/kt/options"
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
	log.Info().Msgf("Cleaning workspace")
	cleanLocalFiles(opts)
	if opts.RuntimeOptions.Component == common.ComponentConnect {
		recoverGlobalHostsAndProxy(opts)
	}

	ctx := context.Background()
	k8s := cli.Kubernetes()
	if opts.RuntimeOptions.Component == common.ComponentExchange {
		recoverExchangedTarget(ctx, opts, k8s)
	} else if opts.RuntimeOptions.Component == common.ComponentMesh {
		recoverAutoMeshRoute(ctx, opts, k8s)
	}
	cleanService(ctx, opts, k8s)
	cleanShadowPodAndConfigMap(ctx, opts, k8s)
}

func recoverGlobalHostsAndProxy(opts *options.DaemonOptions) {
	if strings.HasPrefix(opts.ConnectOptions.DnsMode, common.DnsModeHosts) || opts.ConnectOptions.DnsMode == common.DnsModeLocalDns {
		log.Debug().Msg("Dropping hosts records ...")
		dns.DropHosts()
	}
}

func cleanLocalFiles(opts *options.DaemonOptions) {
	if opts.RuntimeOptions.Component == "" {
		return
	}
	pidFile := fmt.Sprintf("%s/%s-%d.pid", common.KtHome, opts.RuntimeOptions.Component, os.Getpid())
	if err := os.Remove(pidFile); os.IsNotExist(err) {
		log.Debug().Msgf("Pid file %s not exist", pidFile)
	} else if err != nil {
		log.Debug().Err(err).Msgf("Remove pid file %s failed", pidFile)
	} else {
		log.Info().Msgf("Removed pid file %s", pidFile)
	}

	if opts.RuntimeOptions.Shadow != "" {
		for _, sshcm := range strings.Split(opts.RuntimeOptions.Shadow, ",") {
			file := util.PrivateKeyPath(sshcm)
			if err := os.Remove(file); os.IsNotExist(err) {
				log.Debug().Msgf("Key file %s not exist", file)
			} else if err != nil {
				log.Debug().Msgf("Remove key file %s failed", pidFile)
			} else {
				log.Info().Msgf("Removed key file %s", file)
			}
		}
	}
}

func recoverExchangedTarget(ctx context.Context, opts *options.DaemonOptions, k cluster.KubernetesInterface) {
	if opts.RuntimeOptions.Origin == "" {
		// process exit before target exchanged
		return
	}
	if opts.ExchangeOptions.Mode == common.ExchangeModeScale {
		log.Info().Msgf("Recovering origin deployment %s", opts.RuntimeOptions.Origin)
		err := k.ScaleTo(ctx, opts.RuntimeOptions.Origin, opts.Namespace, &opts.RuntimeOptions.Replicas)
		if err != nil {
			log.Error().Err(err).Msgf("Scale deployment %s to %d failed",
				opts.RuntimeOptions.Origin, opts.RuntimeOptions.Replicas)
		}
		// wait for scale complete
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			waitDeploymentRecoverComplete(ctx, opts, k)
			ch <- os.Interrupt
		}()
		_ = <-ch
	} else if opts.ExchangeOptions.Mode == common.ExchangeModeSelector {
		RecoverOriginalService(ctx, k, opts.RuntimeOptions.Origin, opts.Namespace)
	}
}

func recoverAutoMeshRoute(ctx context.Context, opts *options.DaemonOptions, k cluster.KubernetesInterface) {
	if opts.RuntimeOptions.Router != "" {
		routerPod, err := k.GetPod(ctx, opts.RuntimeOptions.Router, opts.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Router pod has been removed unexpectedly")
			return
		}
		if shouldDelRouter, err2 := k.DecreaseRef(ctx, opts.RuntimeOptions.Router, opts.Namespace); err2 != nil {
			log.Error().Err(err2).Msgf("Decrease router pod %s reference failed", opts.RuntimeOptions.Shadow)
		} else if shouldDelRouter {
			recoverService(ctx, k, routerPod.Annotations[common.KtConfig], opts)
		} else {
			stdout, stderr, err3 := k.ExecInPod(common.DefaultContainer, opts.RuntimeOptions.Router, opts.Namespace,
				*opts.RuntimeOptions, common.RouterBin, "remove", opts.RuntimeOptions.Mesh)
			log.Debug().Msgf("Stdout: %s", stdout)
			log.Debug().Msgf("Stderr: %s", stderr)
			if err3 != nil {
				log.Error().Err(err3).Msgf("Failed to remove version %s from router pod", opts.RuntimeOptions.Mesh)
			}
		}
	}
}

func recoverService(ctx context.Context, k cluster.KubernetesInterface, routerConfig string, opts *options.DaemonOptions) {
	config := util.String2Map(routerConfig)
	svcName := config["service"]
	RecoverOriginalService(ctx, k, svcName, opts.Namespace)

	originSvcName := svcName + common.OriginServiceSuffix
	if err := k.RemoveService(ctx, originSvcName, opts.Namespace); err != nil {
		log.Error().Err(err).Msgf("Failed to remove origin service %s", originSvcName)
	}
	log.Info().Msgf("Substitution service %s removed", originSvcName)
}

func RecoverOriginalService(ctx context.Context, k cluster.KubernetesInterface, svcName, namespace string) {
	if svc, err := k.GetService(ctx, svcName, namespace); err != nil {
		log.Error().Err(err).Msgf("Original service %s not found", svcName)
		return
	} else {
		var selector map[string]string
		if svc.Annotations == nil {
			log.Warn().Msgf("No annotation found in service %s, skipping", svcName)
		}
		originSelector, ok := svc.Annotations[common.KtSelector]
		if !ok {
			log.Warn().Msgf("No selector annotation found in service %s, skipping", svcName)
		}
		err = json.Unmarshal([]byte(originSelector), &selector)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to unmarshal original selector of service %s", svcName)
			return
		}
		svc.Spec.Selector = selector
		delete(svc.Annotations, common.KtSelector)
		if _, err = k.UpdateService(ctx, svc); err != nil {
			log.Error().Err(err).Msgf("Failed to recover selector of original service %s", svcName)
		}
	}
	log.Info().Msgf("Original service %s recovered", svcName)
}

func waitDeploymentRecoverComplete(ctx context.Context, opts *options.DaemonOptions, k cluster.KubernetesInterface) {
	ok := false
	counts := opts.ExchangeOptions.RecoverWaitTime / 5
	for i := 0; i < counts; i++ {
		deployment, err := k.GetDeployment(ctx, opts.RuntimeOptions.Origin, opts.Namespace)
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

func cleanService(ctx context.Context, opts *options.DaemonOptions, k cluster.KubernetesInterface) {
	if opts.RuntimeOptions.Service != "" {
		log.Info().Msgf("Cleaning service %s", opts.RuntimeOptions.Service)
		err := k.RemoveService(ctx, opts.RuntimeOptions.Service, opts.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Delete service %s failed", opts.RuntimeOptions.Service)
		}
	}
}

func cleanShadowPodAndConfigMap(ctx context.Context, opts *options.DaemonOptions, k cluster.KubernetesInterface) {
	var err error
	if opts.RuntimeOptions.Shadow != "" {
		shouldDelWithShared := false
		if opts.ConnectOptions.SharedShadow {
			shouldDelWithShared, err = k.DecreaseRef(ctx, opts.RuntimeOptions.Shadow, opts.Namespace)
			if err != nil {
				log.Error().Err(err).Msgf("Decrease shadow daemon pod %s ref count failed", opts.RuntimeOptions.Shadow)
			}
		}
		if shouldDelWithShared || !opts.ConnectOptions.SharedShadow {
			for _, sshcm := range strings.Split(opts.RuntimeOptions.Shadow, ",") {
				log.Info().Msgf("Cleaning configmap %s", sshcm)
				err = k.RemoveConfigMap(ctx, sshcm, opts.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf("Delete configmap %s failed", sshcm)
				}
			}
		}
		if opts.ExchangeOptions != nil && opts.ExchangeOptions.Mode == common.ExchangeModeEphemeral {
			for _, shadow := range strings.Split(opts.RuntimeOptions.Shadow, ",") {
				log.Info().Msgf("Removing ephemeral container of pod %s", shadow)
				err = k.RemoveEphemeralContainer(ctx, common.KtExchangeContainer, shadow, opts.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf("Remove ephemeral container of pod %s failed", shadow)
				}
			}
		} else {
			for _, shadow := range strings.Split(opts.RuntimeOptions.Shadow, ",") {
				log.Info().Msgf("Cleaning shadow pod %s", shadow)
				err = k.RemovePod(ctx, shadow, opts.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf("Delete shadow pod %s failed", shadow)
				}
			}
		}
	}
}
