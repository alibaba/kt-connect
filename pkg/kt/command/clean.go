package command

import (
	"container/list"
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"io/ioutil"
	"k8s.io/api/apps/v1"
	"os"
	"strconv"
	"strings"
	"time"
)

type ResourceToClean struct {
	NamesOfDeploymentToDelete *list.List
	NamesOfServiceToDelete    *list.List
	NamesOfConfigMapToDelete  *list.List
	DeploymentsToScale        map[string]int32
}

// newConnectCommand return new connect command
func newCleanCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "clean",
		Usage: "delete unavailing shadow pods from kubernetes cluster",
		Flags: []urfave.Flag{
			urfave.BoolFlag{
				Name:        "dryRun",
				Usage:       "Only print name of deployments to be deleted",
				Destination: &options.CleanOptions.DryRun,
			},
			urfave.Int64Flag{
				Name:        "thresholdInMinus",
				Usage:       "Length of allowed disconnection time before a unavailing shadow pod be deleted",
				Destination: &options.CleanOptions.ThresholdInMinus,
				Value:       util.ResourceHeartBeatIntervalMinus * 3,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
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
	kubernetes, deployments, err := action.getShadowDeployments(ctx, cli, options)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Found %d shadow deployments", len(deployments))
	resourceToClean := ResourceToClean{list.New(), list.New(), list.New(), make(map[string]int32)}
	for _, deployment := range deployments {
		action.analysisShadowDeployment(deployment, options, resourceToClean)
	}
	if resourceToClean.NamesOfDeploymentToDelete.Len() > 0 {
		if options.CleanOptions.DryRun {
			action.printResourceToClean(resourceToClean)
		} else {
			action.cleanResource(ctx, resourceToClean, kubernetes, options.Namespace)
		}
	} else {
		log.Info().Msg("No unavailing shadow deployment found (^.^)YYa!!")
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
				log.Error().Err(err).
					Msgf("Delete pid file %s failed", f.Name())
			}
		}
	}
}

func (action *Action) analysisShadowDeployment(deployment v1.Deployment, options *options.DaemonOptions, resourceToClean ResourceToClean) {
	lastHeartBeat, err := strconv.ParseInt(deployment.ObjectMeta.Annotations[common.KTLastHeartBeat], 10, 64)
	if err == nil && action.isExpired(lastHeartBeat, options) {
		resourceToClean.NamesOfDeploymentToDelete.PushBack(deployment.Name)
		config := util.String2Map(deployment.ObjectMeta.Annotations[common.KTConfig])
		if deployment.ObjectMeta.Labels[common.KTComponent] == common.ComponentExchange {
			replica, _ := strconv.ParseInt(config["replicas"], 10, 32)
			app := config["app"]
			if replica > 0 && app != "" {
				resourceToClean.DeploymentsToScale[app] = int32(replica)
			}
		} else if deployment.ObjectMeta.Labels[common.KTComponent] == common.ComponentProvide {
			service := config["service"]
			if service != "" {
				resourceToClean.NamesOfServiceToDelete.PushBack(service)
			}
		}
		for _, v := range deployment.Spec.Template.Spec.Volumes {
			if v.ConfigMap != nil && len(v.ConfigMap.Items) == 1 && v.ConfigMap.Items[0].Key == common.SSHAuthKey {
				resourceToClean.NamesOfConfigMapToDelete.PushBack(v.ConfigMap.Name)
			}
		}
	}
}

func (action *Action) cleanResource(ctx context.Context, r ResourceToClean, kubernetes cluster.KubernetesInterface, namespace string) {
	log.Info().Msgf("Deleting %d unavailing shadow deployments", r.NamesOfDeploymentToDelete.Len())
	for name := r.NamesOfDeploymentToDelete.Front(); name != nil; name = name.Next() {
		err := kubernetes.RemoveDeployment(ctx, name.Value.(string), namespace)
		if err != nil {
			log.Error().Msgf("Fail to delete deployment %s", name.Value.(string))
		}
	}
	for name := r.NamesOfServiceToDelete.Front(); name != nil; name = name.Next() {
		err := kubernetes.RemoveService(ctx, name.Value.(string), namespace)
		if err != nil {
			log.Error().Msgf("Fail to delete service %s", name.Value.(string))
		}
	}
	for name := r.NamesOfConfigMapToDelete.Front(); name != nil; name = name.Next() {
		err := kubernetes.RemoveConfigMap(ctx, name.Value.(string), namespace)
		if err != nil {
			log.Error().Msgf("Fail to delete config map %s", name.Value.(string))
		}
	}
	for name, replica := range r.DeploymentsToScale {
		err := kubernetes.ScaleTo(ctx, name, namespace, &replica)
		if err != nil {
			log.Error().Msgf("Fail to scale deployment %s to %d", name, replica)
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
	log.Info().Msgf("Found %d unavailing shadow deployments:", r.NamesOfDeploymentToDelete.Len())
	for name := r.NamesOfDeploymentToDelete.Front(); name != nil; name = name.Next() {
		log.Info().Msgf(" * %s", name.Value.(string))
	}
	log.Info().Msgf("Found %d unavailing shadow service:", r.NamesOfServiceToDelete.Len())
	for name := r.NamesOfServiceToDelete.Front(); name != nil; name = name.Next() {
		log.Info().Msgf(" * %s", name.Value.(string))
	}
	log.Info().Msgf("Found %d exchanged deployments to recover:", len(r.DeploymentsToScale))
	for name, replica := range r.DeploymentsToScale {
		log.Info().Msgf(" * %s -> %d", name, replica)
	}
}

func (action *Action) isExpired(lastHeartBeat int64, options *options.DaemonOptions) bool {
	return time.Now().Unix()-lastHeartBeat > options.CleanOptions.ThresholdInMinus*60
}

func (action *Action) getShadowDeployments(ctx context.Context, cli kt.CliInterface, options *options.DaemonOptions) (
	cluster.KubernetesInterface, []v1.Deployment, error) {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return nil, nil, err
	}
	deployments, err := kubernetes.GetAllExistingShadowDeployments(ctx, options.Namespace)
	if err != nil {
		return nil, nil, err
	}
	return kubernetes, deployments, nil
}
