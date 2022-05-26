package clean

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/dns"
	"github.com/alibaba/kt-connect/pkg/kt/service/tun"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"strings"
)

type ResourceToClean struct {
	PodsToDelete        []string
	ServicesToDelete    []string
	ConfigMapsToDelete  []string
	DeploymentsToDelete []string
	DeploymentsToScale  map[string]int32
	ServicesToRecover   []string
	ServicesToUnlock   []string
}


func CheckClusterResources() (*ResourceToClean, error) {
	pods, cfs, apps, svcs, err := cluster.Ins().GetKtResources(opt.Get().Global.Namespace)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Find %d kt pods", len(pods))
	resourceToClean := ResourceToClean{
		PodsToDelete:        make([]string, 0),
		ServicesToDelete:    make([]string, 0),
		ConfigMapsToDelete:  make([]string, 0),
		DeploymentsToDelete: make([]string, 0),
		DeploymentsToScale:  make(map[string]int32),
		ServicesToRecover:   make([]string, 0),
		ServicesToUnlock:    make([]string, 0),
	}
	for _, pod := range pods {
		analysisExpiredPods(pod, opt.Get().Clean.ThresholdInMinus, &resourceToClean)
	}
	for _, cf := range cfs {
		analysisExpiredConfigmaps(cf, opt.Get().Clean.ThresholdInMinus, &resourceToClean)
	}
	for _, app := range apps {
		analysisExpiredDeployments(app, opt.Get().Clean.ThresholdInMinus, &resourceToClean)
	}
	for _, svc := range svcs {
		analysisExpiredServices(svc, opt.Get().Clean.ThresholdInMinus, &resourceToClean)
	}
	svcList, err := cluster.Ins().GetAllServiceInNamespace(opt.Get().Global.Namespace)
	analysisLockAndOrphanServices(svcList.Items, &resourceToClean)
	return &resourceToClean, nil
}

func TidyClusterResources(r *ResourceToClean) {
	log.Info().Msgf("Deleting %d unavailing kt pods", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		err := cluster.Ins().RemovePod(name, opt.Get().Global.Namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete pods %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing config maps", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		err := cluster.Ins().RemoveConfigMap(name, opt.Get().Global.Namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete config map %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing deployments", len(r.DeploymentsToDelete))
	for _, name := range r.DeploymentsToDelete {
		err := cluster.Ins().RemoveDeployment(name, opt.Get().Global.Namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete deployment %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Recovering %d scaled deployments", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		err := cluster.Ins().ScaleTo(name, opt.Get().Global.Namespace, &replica)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to scale deployment %s to %d", name, replica)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing services", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		err := cluster.Ins().RemoveService(name, opt.Get().Global.Namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete service %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Recovering %d meshed services", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		general.RecoverOriginalService(name, opt.Get().Global.Namespace)
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Recovering %d locked services", len(r.ServicesToUnlock))
	for _, name := range r.ServicesToUnlock {
		if app, err := cluster.Ins().GetService(name, opt.Get().Global.Namespace); err == nil {
			delete(app.Annotations, util.KtLock)
			_, err = cluster.Ins().UpdateService(app)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to lock service %s", name)
			} else {
				log.Info().Msgf(" * %s", name)
			}
		}
	}
	log.Info().Msg("Done")
}

func PrintClusterResourcesToClean(r *ResourceToClean) {
	log.Info().Msgf("Find %d unavailing pods to delete:", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d unavailing config maps to delete:", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d unavailing deployments to delete:", len(r.DeploymentsToDelete))
	for _, name := range r.DeploymentsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d exchanged deployments to recover:", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		log.Info().Msgf(" * %s -> %d", name, replica)
	}
	log.Info().Msgf("Find %d unavailing service to delete:", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d meshed service to recover:", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d locked services to recover:", len(r.ServicesToUnlock))
	for _, name := range r.ServicesToUnlock {
		log.Info().Msgf(" * %s", name)
	}
}

func TidyLocalResources() {
	log.Debug().Msg("Cleaning up unused pid files")
	cleanPidFiles()
	log.Debug().Msg("Cleaning up unused local rsa keys")
	util.CleanRsaKeys()
	log.Debug().Msg("Cleaning up background logs")
	util.CleanBackgroundLogs()
	if util.GetDaemonRunning(util.ComponentConnect) < 0 {
		if util.IsRunAsAdmin() {
			log.Debug().Msg("Cleaning up hosts file")
			dns.DropHosts()
			log.Debug().Msg("Cleaning DNS configuration")
			dns.Ins().RestoreNameServer()
			log.Info().Msgf("Cleaning route table")
			if err := tun.Ins().RestoreRoute(); err != nil {
				log.Warn().Err(err).Msgf("Unable to clean up route table")
			}
		} else {
			log.Info().Msgf("Not %s user, DNS cleanup skipped", util.GetAdminUserName())
		}
	}
}

func cleanPidFiles() {
	files, _ := ioutil.ReadDir(util.KtPidDir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pid") {
			component, pid := parseComponentAndPid(f.Name())
			if util.IsProcessExist(pid) {
				log.Debug().Msgf("Find kt %s instance with pid %d", component, pid)
			} else {
				log.Info().Msgf("Removing remnant pid file %s", f.Name())
				if err := os.Remove(fmt.Sprintf("%s/%s", util.KtPidDir, f.Name())); err != nil {
					log.Error().Err(err).Msgf("Delete pid file %s failed", f.Name())
				}
			}
		}
	}
}

func parseComponentAndPid(pidFileName string) (string, int) {
	startPos := strings.LastIndex(pidFileName, "-")
	endPos := strings.Index(pidFileName, ".")
	if startPos > 0 && endPos > startPos {
		component := pidFileName[0 : startPos]
		pid, err := strconv.Atoi(pidFileName[startPos+1 : endPos])
		if err != nil {
			return "", -1
		}
		return component, pid
	}
	return "", -1
}

func analysisExpiredPods(pod coreV1.Pod, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(pod.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat < 0 {
		log.Debug().Msgf("Pod %s does no have heart beat annotation", pod.Name)
	} else if isExpired(lastHeartBeat, cleanThresholdInMinus) {
		log.Debug().Msgf(" * pod %s expired, lastHeartBeat: %d ", pod.Name, lastHeartBeat)
		if pod.DeletionTimestamp == nil {
			resourceToClean.PodsToDelete = append(resourceToClean.PodsToDelete, pod.Name)
		}
		analysisConfigAnnotation(pod.Labels[util.KtRole], util.String2Map(pod.Annotations[util.KtConfig]), resourceToClean)
	}
}

func analysisExpiredConfigmaps(cf coreV1.ConfigMap, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(cf.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat < 0 {
		log.Debug().Msgf("Configmap %s does no have heart beat annotation", cf.Name)
	} else if isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.ConfigMapsToDelete = append(resourceToClean.ConfigMapsToDelete, cf.Name)
	}
}

func analysisExpiredDeployments(app appV1.Deployment, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(app.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat < 0 {
		log.Debug().Msgf("Deployment %s does no have heart beat annotation", app.Name)
	} else if isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.DeploymentsToDelete = append(resourceToClean.DeploymentsToDelete, app.Name)
		analysisConfigAnnotation(app.Labels[util.KtRole], util.String2Map(app.Annotations[util.KtConfig]), resourceToClean)
	}
}

func analysisExpiredServices(svc coreV1.Service, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(svc.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat < 0 {
		log.Debug().Msgf("Service %s does no have heart beat annotation", svc.Name)
	} else if isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, svc.Name)
	}
}

func analysisLockAndOrphanServices(svcs []coreV1.Service, resourceToClean *ResourceToClean) {
	for _, svc := range svcs {
		if svc.Annotations == nil {
			continue
		}
		if lock, exists := svc.Annotations[util.KtLock]; exists && util.GetTime() - util.ParseTimestamp(lock) > general.LockTimeout {
			resourceToClean.ServicesToUnlock = append(resourceToClean.ServicesToUnlock, svc.Name)
		}
		if svc.Annotations[util.KtSelector] != "" {
			if svc.Spec.Selector[util.KtRole] == util.RoleRouter {
				// it's a meshed service, but router pod already gone
				if !isRouterPodExist(svc.Name, svc.Namespace) {
					resourceToClean.ServicesToRecover = append(resourceToClean.ServicesToRecover, svc.Name)
				}
			} else {
				// it's an exchanged service, but shadow pod already gone
				if !isShadowPodExist(svc.Spec.Selector, svc.Name, svc.Namespace, util.KtExchangeContainer) {
					resourceToClean.ServicesToRecover = append(resourceToClean.ServicesToRecover, svc.Name)
				}
			}
		}
	}
}

func analysisConfigAnnotation(role string, config map[string]string, resourceToClean *ResourceToClean) {
	log.Debug().Msgf("   role %s, config: %v", role, config)
	// scale exchange
	if role == util.RoleExchangeShadow {
		replica, _ := strconv.ParseInt(config["replicas"], 10, 32)
		app := config["app"]
		if replica > 0 && app != "" {
			resourceToClean.DeploymentsToScale[app] = int32(replica)
		}
	}
	// auto mesh and selector exchange
	if role == util.RoleRouter || role == util.RoleExchangeShadow {
		if service, exists := config["service"]; exists {
			resourceToClean.ServicesToRecover = append(resourceToClean.ServicesToRecover, service)
		}
	}
}

func isShadowPodExist(selector map[string]string, svcName, namespace, suffix string) bool {
	pods, err := cluster.Ins().GetPodsByLabel(selector, namespace)
	if err != nil {
		return false
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, fmt.Sprintf("%s-%s-", svcName, suffix)) {
			return true
		}
	}
	return false
}

func isRouterPodExist(svcName, namespace string) bool {
	routerPodName := svcName + util.RouterPodSuffix
	_, err := cluster.Ins().GetPod(routerPodName, namespace)
	return err == nil
}

func isExpired(lastHeartBeat, cleanThresholdInMinus int64) bool {
	return util.GetTime() - lastHeartBeat > cleanThresholdInMinus*60
}

