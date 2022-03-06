package clean

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
	"time"
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

func AnalysisExpiredPods(pod coreV1.Pod, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(pod.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat > 0 && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		log.Debug().Msgf(" * pod %s expired, lastHeartBeat: %d ", pod.Name, lastHeartBeat)
		if pod.DeletionTimestamp == nil {
			resourceToClean.PodsToDelete = append(resourceToClean.PodsToDelete, pod.Name)
		}
		analysisConfigAnnotation(pod.Labels[util.KtRole], util.String2Map(pod.Annotations[util.KtConfig]), resourceToClean)
	} else {
		log.Debug().Msgf("Pod %s does no have heart beat annotation", pod.Name)
	}
}

func AnalysisExpiredConfigmaps(cf coreV1.ConfigMap, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(cf.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat > 0 && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.ConfigMapsToDelete = append(resourceToClean.ConfigMapsToDelete, cf.Name)
	}
}

func AnalysisExpiredDeployments(app appV1.Deployment, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(app.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat > 0 && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.DeploymentsToDelete = append(resourceToClean.DeploymentsToDelete, app.Name)
		analysisConfigAnnotation(app.Labels[util.KtRole], util.String2Map(app.Annotations[util.KtConfig]), resourceToClean)
	}
}

func AnalysisExpiredServices(svc coreV1.Service, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat := util.ParseTimestamp(svc.Annotations[util.KtLastHeartBeat])
	if lastHeartBeat > 0 && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, svc.Name)
	}
}

func AnalysisLockAndOrphanServices(svcs []coreV1.Service, resourceToClean *ResourceToClean) {
	for _, svc := range svcs {
		if svc.Annotations == nil {
			continue
		}
		if lock, ok := svc.Annotations[util.KtLock]; ok && time.Now().Unix() - util.ParseTimestamp(lock) > general.LockTimeout {
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
		if service, ok := config["service"]; ok {
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

func TidyResource(r ResourceToClean, namespace string) {
	log.Info().Msgf("Deleting %d unavailing kt pods", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		err := cluster.Ins().RemovePod(name, namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete pods %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing config maps", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		err := cluster.Ins().RemoveConfigMap(name, namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete config map %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing deployments", len(r.DeploymentsToDelete))
	for _, name := range r.DeploymentsToDelete {
		err := cluster.Ins().RemoveDeployment(name, namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete deployment %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Recovering %d scaled deployments", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		err := cluster.Ins().ScaleTo(name, namespace, &replica)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to scale deployment %s to %d", name, replica)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing services", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		err := cluster.Ins().RemoveService(name, namespace)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to delete service %s", name)
		} else {
			log.Info().Msgf(" * %s", name)
		}
	}
	log.Info().Msgf("Recovering %d meshed services", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		general.RecoverOriginalService(name, namespace)
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Recovering %d locked services", len(r.ServicesToUnlock))
	for _, name := range r.ServicesToUnlock {
		if app, err := cluster.Ins().GetService(name, namespace); err == nil {
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

func PrintResourceToClean(r ResourceToClean) {
	log.Info().Msgf("Find %d unavailing pods to delete:", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d unavailing config maps to delete:", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Find %d unavailing deployments to delete:", len(r.DeploymentsToScale))
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

func isExpired(lastHeartBeat, cleanThresholdInMinus int64) bool {
	return time.Now().Unix()-lastHeartBeat > cleanThresholdInMinus*60
}

func IsEmpty(r ResourceToClean) bool {
	return len(r.PodsToDelete) == 0 &&
		len(r.ConfigMapsToDelete) == 0 &&
		len(r.DeploymentsToDelete) == 0 &&
		len(r.DeploymentsToScale) == 0 &&
		len(r.ServicesToDelete) == 0 &&
		len(r.ServicesToUnlock) == 0 &&
		len(r.ServicesToRecover) == 0
}
