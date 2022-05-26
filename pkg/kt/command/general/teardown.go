package general

import (
	"encoding/json"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/dns"
	"github.com/alibaba/kt-connect/pkg/kt/service/tun"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// CleanupWorkspace clean workspace
func CleanupWorkspace() {
	log.Debug().Msgf("Cleaning workspace")
	cleanLocalFiles()
	if opt.Store.Component == util.ComponentConnect {
		recoverGlobalHostsAndProxy()
	}

	if opt.Store.Component == util.ComponentExchange {
		recoverExchangedTarget()
	} else if opt.Store.Component == util.ComponentMesh {
		recoverAutoMeshRoute()
	}
	cleanService()
	cleanShadowPodAndConfigMap()
}

func recoverGlobalHostsAndProxy() {
	if strings.HasPrefix(opt.Get().Connect.DnsMode, util.DnsModeHosts) ||
		strings.HasPrefix(opt.Get().Connect.DnsMode, util.DnsModeLocalDns) {
		log.Debug().Msg("Dropping hosts records ...")
		dns.DropHosts()
	}
	if strings.HasPrefix(opt.Get().Connect.DnsMode, util.DnsModeLocalDns) {
		if err := tun.Ins().RestoreRoute(); err != nil {
			log.Debug().Err(err).Msgf("Failed to restore route table")
		}
	}
}

func cleanLocalFiles() {
	if opt.Store.Component == "" {
		return
	}
	pidFile := fmt.Sprintf("%s/%s-%d.pid", util.KtPidDir, opt.Store.Component, os.Getpid())
	if err := os.Remove(pidFile); os.IsNotExist(err) {
		log.Debug().Msgf("Pid file %s not exist", pidFile)
	} else if err != nil {
		log.Debug().Err(err).Msgf("Remove pid file %s failed", pidFile)
	} else {
		log.Info().Msgf("Removed pid file %s", pidFile)
	}

	if opt.Store.Shadow != "" {
		for _, sshcm := range strings.Split(opt.Store.Shadow, ",") {
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

func recoverExchangedTarget() {
	if opt.Store.Origin == "" {
		// process exit before target exchanged
		return
	}
	if opt.Get().Exchange.Mode == util.ExchangeModeScale {
		log.Info().Msgf("Recovering origin deployment %s", opt.Store.Origin)
		err := cluster.Ins().ScaleTo(opt.Store.Origin, opt.Get().Global.Namespace, &opt.Store.Replicas)
		if err != nil {
			log.Error().Err(err).Msgf("Scale deployment %s to %d failed",
				opt.Store.Origin, opt.Store.Replicas)
		}
		// wait for scale complete
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		go func() {
			waitDeploymentRecoverComplete()
			ch <- os.Interrupt
		}()
		_ = <-ch
	} else if opt.Get().Exchange.Mode == util.ExchangeModeSelector {
		RecoverOriginalService(opt.Store.Origin, opt.Get().Global.Namespace)
		log.Info().Msgf("Original service %s recovered", opt.Store.Origin)
	}
}

func recoverAutoMeshRoute() {
	if opt.Store.Router != "" {
		routerPod, err := cluster.Ins().GetPod(opt.Store.Router, opt.Get().Global.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Router pod has been removed unexpectedly")
			// in case of router pod gone, try recover origin service via runtime store
			if opt.Store.Origin != "" {
				recoverService(opt.Store.Origin)
			}
			return
		}
		if shouldDelRouter, err2 := cluster.Ins().DecreasePodRef(opt.Store.Router, opt.Get().Global.Namespace); err2 != nil {
			log.Error().Err(err2).Msgf("Decrease router pod %s reference failed", opt.Store.Shadow)
		} else if shouldDelRouter {
			routerConfig := routerPod.Annotations[util.KtConfig]
			config := util.String2Map(routerConfig)
			recoverService(config["service"])
			if err = cluster.Ins().RemovePod(opt.Store.Router, opt.Get().Global.Namespace); err != nil {
				log.Warn().Err(err).Msgf("Failed to remove router pod")
			}
		} else {
			stdout, stderr, err3 := cluster.Ins().ExecInPod(util.DefaultContainer, opt.Store.Router, opt.Get().Global.Namespace,
				util.RouterBin, "remove", opt.Store.Mesh)
			log.Debug().Msgf("Stdout: %s", stdout)
			log.Debug().Msgf("Stderr: %s", stderr)
			if err3 != nil {
				log.Warn().Err(err3).Msgf("Failed to remove version %s from router pod", opt.Store.Mesh)
			}
		}
	}
}

func recoverService(originSvcName string) {
	RecoverOriginalService(originSvcName, opt.Get().Global.Namespace)
	log.Info().Msgf("Original service %s recovered", originSvcName)

	stuntmanSvcName := originSvcName + util.StuntmanServiceSuffix
	if err := cluster.Ins().RemoveService(stuntmanSvcName, opt.Get().Global.Namespace); err != nil {
		log.Error().Err(err).Msgf("Failed to remove stuntman service %s", stuntmanSvcName)
	}
	log.Info().Msgf("Stuntman service %s removed", stuntmanSvcName)
}

func RecoverOriginalService(svcName, namespace string) {
	if svc, err := cluster.Ins().GetService(svcName, namespace); err != nil {
		log.Error().Err(err).Msgf("Original service %s not found", svcName)
		return
	} else {
		var selector map[string]string
		if svc.Annotations == nil {
			log.Warn().Msgf("No annotation found in service %s, skipping", svcName)
			return
		}
		originSelector, exists := svc.Annotations[util.KtSelector]
		if !exists {
			log.Warn().Msgf("No selector annotation found in service %s, skipping", svcName)
			return
		}
		if err = json.Unmarshal([]byte(originSelector), &selector); err != nil {
			log.Error().Err(err).Msgf("Failed to unmarshal original selector of service %s", svcName)
			return
		}
		svc.Spec.Selector = selector
		delete(svc.Annotations, util.KtSelector)
		if _, err = cluster.Ins().UpdateService(svc); err != nil {
			log.Error().Err(err).Msgf("Failed to recover selector of original service %s", svcName)
		}
	}
}

func waitDeploymentRecoverComplete() {
	ok := false
	counts := opt.Get().Exchange.RecoverWaitTime / 5
	for i := 0; i < counts; i++ {
		deployment, err := cluster.Ins().GetDeployment(opt.Store.Origin, opt.Get().Global.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Cannot fetch original deployment %s", opt.Store.Origin)
			break
		} else if deployment.Status.ReadyReplicas == opt.Store.Replicas {
			ok = true
			break
		} else {
			log.Info().Msgf("Wait for deployment %s recover ...", opt.Store.Origin)
			time.Sleep(5 * time.Second)
		}
	}
	if !ok {
		log.Warn().Msgf("Deployment %s recover timeout", opt.Store.Origin)
	}
}

func cleanService() {
	if opt.Store.Service != "" {
		log.Info().Msgf("Cleaning service %s", opt.Store.Service)
		err := cluster.Ins().RemoveService(opt.Store.Service, opt.Get().Global.Namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Delete service %s failed", opt.Store.Service)
		}
	}
}

func cleanShadowPodAndConfigMap() {
	var err error
	if opt.Store.Shadow != "" {
		shouldDelWithShared := false
		if opt.Get().Connect.ShareShadow {
			// There is always exactly one shadow pod or deployment for connect
			if opt.Get().Global.UseShadowDeployment {
				shouldDelWithShared, err = cluster.Ins().DecreaseDeploymentRef(opt.Store.Shadow, opt.Get().Global.Namespace)
			} else {
				shouldDelWithShared, err = cluster.Ins().DecreasePodRef(opt.Store.Shadow, opt.Get().Global.Namespace)
			}
			if err != nil {
				log.Error().Err(err).Msgf("Decrease shadow daemon %s ref count failed", opt.Store.Shadow)
			}
		}
		if shouldDelWithShared || !opt.Get().Connect.ShareShadow {
			for _, shadow := range strings.Split(opt.Store.Shadow, ",") {
				log.Info().Msgf("Cleaning configmap %s", shadow)
				err = cluster.Ins().RemoveConfigMap(shadow, opt.Get().Global.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf("Delete configmap %s failed", shadow)
				}
				log.Info().Msgf("Cleaning shadow pod %s", shadow)
				if opt.Get().Global.UseShadowDeployment {
					err = cluster.Ins().RemoveDeployment(shadow, opt.Get().Global.Namespace)
				} else {
					err = cluster.Ins().RemovePod(shadow, opt.Get().Global.Namespace)
				}
				if err != nil {
					log.Error().Err(err).Msgf("Delete shadow pod %s failed", shadow)
				}
			}
		}
		if opt.Get().Exchange.Mode == util.ExchangeModeEphemeral {
			for _, shadow := range strings.Split(opt.Store.Shadow, ",") {
				log.Info().Msgf("Removing ephemeral container of pod %s", shadow)
				err = cluster.Ins().RemoveEphemeralContainer(util.KtExchangeContainer, shadow, opt.Get().Global.Namespace)
				if err != nil {
					log.Error().Err(err).Msgf("Remove ephemeral container of pod %s failed", shadow)
				}
			}
		}
	}
}
