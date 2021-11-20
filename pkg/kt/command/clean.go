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
	coreV1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"strings"
	"time"
)

type ResourceToClean struct {
	PodsToDelete       []string
	ServicesToDelete   []string
	ConfigMapsToDelete []string
	DeploymentsToScale map[string]int32
	ServicesToRecover  []string
}

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

	kubernetes, pods, err := action.getShadowAndRouterPods(ctx, cli, options)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Found %d kt pods", len(pods))
	resourceToClean := ResourceToClean{make([]string, 0), make([]string, 0), make([]string, 0), make(map[string]int32), make([]string, 0)}
	for _, pod := range pods {
		action.analysisExpiredPods(pod, options, &resourceToClean)
	}
	if len(resourceToClean.PodsToDelete) > 0 {
		if options.CleanOptions.DryRun {
			action.printResourceToClean(resourceToClean)
		} else {
			action.cleanResource(ctx, resourceToClean, kubernetes, options.Namespace)
		}
	} else {
		log.Info().Msg("No unavailing kt pod found (^.^)YYa!!")
	}

	// TODO: unlock mesh deployments

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

func (action *Action) analysisExpiredPods(pod coreV1.Pod, options *options.DaemonOptions, resourceToClean *ResourceToClean) {
	lastHeartBeat, err := strconv.ParseInt(pod.Annotations[common.KtLastHeartBeat], 10, 64)
	if err == nil && action.isExpired(lastHeartBeat, options) {
		log.Debug().Msgf(" * pod %s expired, lastHeartBeat: %d ", pod.Name, lastHeartBeat)
		resourceToClean.PodsToDelete = append(resourceToClean.PodsToDelete, pod.Name)
		log.Debug().Msgf("   component %s, config: %s", pod.Labels[common.KtComponent], pod.Annotations[common.KtConfig])
		config := util.String2Map(pod.Annotations[common.KtConfig])
		if pod.Labels[common.KtComponent] == common.ComponentExchange {
			replica, _ := strconv.ParseInt(config["replicas"], 10, 32)
			app := config["app"]
			if replica > 0 && app != "" {
				resourceToClean.DeploymentsToScale[app] = int32(replica)
			}
		} else if pod.Labels[common.KtComponent] == common.ComponentProvide {
			if service, ok := config["service"]; ok {
				resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, service)
			}
		} else if pod.Labels[common.KtComponent] == common.ComponentMesh {
			if service, ok := config["service"]; ok {
				switch pod.Labels[common.KtRole] {
				case common.RoleShadow:
					resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, service)
				case common.RoleRouter:
					resourceToClean.ServicesToRecover = append(resourceToClean.ServicesToRecover, service)
					resourceToClean.ServicesToDelete = append(resourceToClean.ServicesToDelete, service + common.OriginServiceSuffix)
				}
			}
		}
		for _, v := range pod.Spec.Volumes {
			if v.ConfigMap != nil && len(v.ConfigMap.Items) == 1 && v.ConfigMap.Items[0].Key == common.SshAuthKey {
				resourceToClean.ConfigMapsToDelete = append(resourceToClean.ConfigMapsToDelete, v.ConfigMap.Name)
			}
		}
	} else {
		log.Debug().Msgf("Pod %s does no have heart beat annotation")
	}
}

func (action *Action) cleanResource(ctx context.Context, r ResourceToClean, k cluster.KubernetesInterface, namespace string) {
	log.Info().Msgf("Deleting %d unavailing kt pods", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		err := k.RemovePod(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Fail to delete pods %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing servicse", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		err := k.RemoveService(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Fail to delete service %s", name)
		}
	}
	log.Info().Msgf("Deleting %d unavailing config maps", len(r.ConfigMapsToDelete))
	for _, name := range r.ConfigMapsToDelete {
		err := k.RemoveConfigMap(ctx, name, namespace)
		if err != nil {
			log.Error().Err(err).Msgf("Fail to delete config map %s", name)
		}
	}
	log.Info().Msgf("Recovering %d scaled deployments", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		err := k.ScaleTo(ctx, name, namespace, &replica)
		if err != nil {
			log.Error().Err(err).Msgf("Fail to scale deployment %s to %d", name, replica)
		}
	}
	for _, name := range r.ServicesToRecover {
		general.RecoverOriginalService(ctx, k, name, namespace)
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
	log.Info().Msgf("Found %d unavailing pods:", len(r.PodsToDelete))
	for _, name := range r.PodsToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d unavailing service:", len(r.ServicesToDelete))
	for _, name := range r.ServicesToDelete {
		log.Info().Msgf(" * %s", name)
	}
	log.Info().Msgf("Found %d exchanged deployments to recover:", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		log.Info().Msgf(" * %s -> %d", name, replica)
	}
	log.Info().Msgf("Found %d meshed service to recover:", len(r.ServicesToRecover))
	for _, name := range r.ServicesToRecover {
		log.Info().Msgf(" * %s", name)
	}
}

func (action *Action) isExpired(lastHeartBeat int64, options *options.DaemonOptions) bool {
	return time.Now().Unix()-lastHeartBeat > options.CleanOptions.ThresholdInMinus*60
}

func (action *Action) getShadowAndRouterPods(ctx context.Context, cli kt.CliInterface, options *options.DaemonOptions) (
	cluster.KubernetesInterface, []coreV1.Pod, error) {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return nil, nil, err
	}
	pods, err := cluster.GetKtPods(ctx, kubernetes, options.Namespace)
	if err != nil {
		return nil, nil, err
	}
	return kubernetes, pods, nil
}
