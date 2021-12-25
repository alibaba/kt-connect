package command

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"io/ioutil"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"strings"
	"time"
)

type ResourceToClean struct {
	PodsToDelete        []string
	ServicesToDelete    []string
	ConfigMapsToDelete  []string
	DeploymentsToScale  map[string]int32
	DeploymentsToUnlock []string
	ServicesToRecover   []string
	ServicesToUnlock   []string
}

const SecondsOf10Mins = 10 * 60

// NewCleanCommand return new connect command
func NewCleanCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "clean",
		Usage: "delete unavailing shadow pods from kubernetes cluster",
		Flags: general.CleanActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
				return err
			}
			return action.Clean(cli, options)
		},
	}
}

//Clean delete unavailing shadow pods
func (action *Action) Clean(cli kt.CliInterface, options *options.DaemonOptions) error {
	action.cleanPidFiles()
	ctx := context.Background()

	pods, deployments, svcs, err := cluster.GetKtResources(ctx, cli.Kubernetes(), options.Namespace)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Found %d kt pods", len(pods))
	resourceToClean := ResourceToClean{
		PodsToDelete: make([]string, 0),
		ServicesToDelete: make([]string, 0),
		ConfigMapsToDelete: make([]string, 0),
		DeploymentsToScale: make(map[string]int32),
		DeploymentsToUnlock: make([]string, 0),
		ServicesToRecover: make([]string, 0),
		ServicesToUnlock: make([]string, 0),
	}
	for _, pod := range pods {
		action.analysisExpiredPods(pod, options.CleanOptions.ThresholdInMinus, &resourceToClean)
	}
	for _, svc := range svcs {
		action.analysisExpiredServices(svc, options.CleanOptions.ThresholdInMinus, &resourceToClean)
	}
	action.analysisLocked(deployments, svcs, &resourceToClean)
	if isEmpty(resourceToClean) {
		log.Info().Msg("No unavailing kt resource found (^.^)YYa!!")
	} else {
		if options.CleanOptions.DryRun {
			action.printResourceToClean(resourceToClean)
		} else {
			action.cleanResource(ctx, resourceToClean, cli.Kubernetes(), options.Namespace)
		}
	}

	if !options.CleanOptions.DryRun {
		log.Debug().Msg("Cleaning up unused local rsa keys ...")
		util.CleanRsaKeys()
		log.Debug().Msg("Cleaning up hosts file ...")
		util.DropHosts()
		log.Debug().Msg("Cleaning up global proxy and environment variable ...")
		registry.ResetGlobalProxyAndEnvironmentVariable()
	}
	return nil
}

func (action *Action) cleanPidFiles() {
	files, _ := ioutil.ReadDir(util.KtHome)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pid") && !util.IsProcessExist(action.toPid(f.Name())) {
			log.Info().Msgf("Removing pid file %s", f.Name())
			if err := os.Remove(fmt.Sprintf("%s/%s", util.KtHome, f.Name())); err != nil {
				log.Error().Err(err).Msgf("Delete pid file %s failed", f.Name())
			}
		}
	}
}

func (action *Action) analysisExpiredPods(pod coreV1.Pod, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat, err := strconv.ParseInt(pod.Annotations[common.KtLastHeartBeat], 10, 64)
	if err == nil && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		log.Debug().Msgf(" * pod %s expired, lastHeartBeat: %d ", pod.Name, lastHeartBeat)
		resourceToClean.PodsToDelete = append(resourceToClean.PodsToDelete, pod.Name)
		log.Debug().Msgf("   role %s, config: %s", pod.Labels[common.KtRole], pod.Annotations[common.KtConfig])
		config := util.String2Map(pod.Annotations[common.KtConfig])
		if pod.Labels[common.KtRole] == common.RoleExchangeShadow {
			replica, _ := strconv.ParseInt(config["replicas"], 10, 32)
			app := config["app"]
			if replica > 0 && app != "" {
				resourceToClean.DeploymentsToScale[app] = int32(replica)
			}
		} else if pod.Labels[common.KtRole] == common.RoleProvideShadow || pod.Labels[common.KtRole] == common.RoleMeshShadow {
			if service, ok := config["service"]; ok {
				resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, service)
			}
		} else if pod.Labels[common.KtRole] == common.RoleRouter {
			if service, ok := config["service"]; ok {
				resourceToClean.ServicesToRecover = append(resourceToClean.ServicesToRecover, service)
				resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, service + common.OriginServiceSuffix)
			}
		}
		for _, v := range pod.Spec.Volumes {
			if v.ConfigMap != nil && len(v.ConfigMap.Items) == 1 && v.ConfigMap.Items[0].Key == common.SshAuthKey {
				resourceToClean.ConfigMapsToDelete = append(resourceToClean.ConfigMapsToDelete, v.ConfigMap.Name)
			}
		}
	} else {
		log.Debug().Msgf("Pod %s does no have heart beat annotation", pod.Name)
	}
}

func (action *Action) analysisExpiredServices(svc coreV1.Service, cleanThresholdInMinus int64, resourceToClean *ResourceToClean) {
	lastHeartBeat, err := strconv.ParseInt(svc.Annotations[common.KtLastHeartBeat], 10, 64)
	if err == nil && isExpired(lastHeartBeat, cleanThresholdInMinus) {
		resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, svc.Name)
	}
}

func (action *Action) analysisLocked(apps []appV1.Deployment, svcs []coreV1.Service, resourceToClean *ResourceToClean) {
	for _, app := range apps {
		if lock, ok := app.Annotations[common.KtLock]; ok {
			lockTime, err := strconv.ParseInt(lock, 10, 64)
			if err == nil && time.Now().Unix() - lockTime > SecondsOf10Mins {
				resourceToClean.DeploymentsToUnlock = append(resourceToClean.DeploymentsToUnlock, app.Name)
			}
		}
	}
	for _, svc := range svcs {
		if lock, ok := svc.Annotations[common.KtLock]; ok {
			lockTime, err := strconv.ParseInt(lock, 10, 64)
			if err == nil && time.Now().Unix() - lockTime > SecondsOf10Mins {
				resourceToClean.ServicesToUnlock = append(resourceToClean.ServicesToUnlock, svc.Name)
			}
		}
	}
}

func (action *Action) cleanResource(ctx context.Context, r ResourceToClean, k cluster.KubernetesInterface, namespace string) {
	log.Info().Msgf("Deleting %d unavailing kt pods", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		err := k.RemovePod(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to delete pods %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing config maps", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		err := k.RemoveConfigMap(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to delete config map %s", name)
		}
	}
	log.Info().Msgf("Recovering %d scaled deployments", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		err := k.ScaleTo(ctx, name, namespace, &replica)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to scale deployment %s to %d", name, replica)
		}
	}
	log.Info().Msgf("Recovering %d locked deployments", len(r.DeploymentsToUnlock))
	for _, name := range r.DeploymentsToUnlock {
		if app, err := k.GetDeployment(ctx, name, namespace); err == nil {
			delete(app.Annotations, common.KtLock)
			_, err = k.UpdateDeployment(ctx, app)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to lock deployment %s", name)
			}
		}
	}
	log.Info().Msgf("Deleting %d unavailing services", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		err := k.RemoveService(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to delete service %s", name)
		}
	}
	log.Info().Msgf("Recovering %d meshed services", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		general.RecoverOriginalService(ctx, k, name, namespace)
	}
	log.Info().Msgf("Recovering %d locked services", len(r.ServicesToUnlock))
	for _, name := range r.ServicesToUnlock {
		if app, err := k.GetService(ctx, name, namespace); err == nil {
			delete(app.Annotations, common.KtLock)
			_, err = k.UpdateService(ctx, app)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to lock service %s", name)
			}
		}
	}
	log.Info().Msg("Done")
}

func (action *Action) toPid(pidFileName string) int {
	startPos := strings.LastIndex(pidFileName, "-")
	endPos := strings.Index(pidFileName, ".")
	if startPos > 0 && endPos > startPos {
		pid, err := strconv.Atoi(pidFileName[startPos+1 : endPos])
		if err != nil {
			return -1
		}
		return pid
	}
	return -1
}

func (action *Action) printResourceToClean(r ResourceToClean) {
	log.Info().Msgf("Found %d unavailing pods to delete:", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d unavailing config maps to delete:", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d exchanged deployments to recover:", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		log.Info().Msgf(" * %s -> %d", name, replica)
	}
	log.Info().Msgf("Found %d locked deployments to recover:", len(r.DeploymentsToUnlock))
	for _, name := range r.DeploymentsToUnlock {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d unavailing service to delete:", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d meshed service to recover:", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d locked services to recover:", len(r.ServicesToUnlock))
	for _, name := range r.ServicesToUnlock {
		log.Info().Msgf(" * %s", name)
	}
}

func isExpired(lastHeartBeat, cleanThresholdInMinus int64) bool {
	return time.Now().Unix()-lastHeartBeat > cleanThresholdInMinus*60
}

func isEmpty(r ResourceToClean) bool {
	return len(r.ServicesToDelete) == 0 &&
		len(r.PodsToDelete) == 0 &&
		len(r.ConfigMapsToDelete) == 0 &&
		len(r.ServicesToUnlock) == 0 &&
		len(r.DeploymentsToUnlock) == 0 &&
		len(r.ServicesToRecover) == 0 &&
		len(r.DeploymentsToScale) == 0
}
