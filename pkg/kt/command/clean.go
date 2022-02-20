package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/dns"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// NewCleanCommand return new connect command
func NewCleanCommand(action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "clean",
		Usage: "delete unavailing resources created by kt from kubernetes cluster",
		UsageText: "ktctl clean [command options]",
		Flags: general.CleanActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}
			return action.Clean()
		},
	}
}

// Clean delete unavailing shadow pods
func (action *Action) Clean() error {
	cleanPidFiles()
	pods, cfs, svcs, err := cluster.Ins().GetKtResources(opt.Get().Namespace)
	if err != nil {
		return err
	}
	log.Debug().Msgf("Find %d kt pods", len(pods))
	resourceToClean := clean.ResourceToClean{
		PodsToDelete: make([]string, 0),
		ServicesToDelete: make([]string, 0),
		ConfigMapsToDelete: make([]string, 0),
		DeploymentsToScale: make(map[string]int32),
		ServicesToRecover: make([]string, 0),
		ServicesToUnlock: make([]string, 0),
	}
	for _, pod := range pods {
		clean.AnalysisExpiredPods(pod, opt.Get().CleanOptions.ThresholdInMinus, &resourceToClean)
	}
	for _, cf := range cfs {
		clean.AnalysisExpiredConfigmaps(cf, opt.Get().CleanOptions.ThresholdInMinus, &resourceToClean)
	}
	for _, svc := range svcs {
		clean.AnalysisExpiredServices(svc, opt.Get().CleanOptions.ThresholdInMinus, &resourceToClean)
	}
	svcList, err := cluster.Ins().GetAllServiceInNamespace(opt.Get().Namespace)
	clean.AnalysisLockAndOrphanServices(svcList.Items, &resourceToClean)
	if clean.IsEmpty(resourceToClean) {
		log.Info().Msg("No unavailing kt resource found (^.^)YYa!!")
	} else {
		if opt.Get().CleanOptions.DryRun {
			clean.PrintResourceToClean(resourceToClean)
		} else {
			clean.TidyResource(resourceToClean, opt.Get().Namespace)
		}
	}

	if !opt.Get().CleanOptions.DryRun {
		log.Debug().Msg("Cleaning up unused local rsa keys ...")
		util.CleanRsaKeys()
		if util.GetDaemonRunning(util.ComponentConnect) < 0 {
			if util.IsRunAsAdmin() {
				log.Debug().Msg("Cleaning up hosts file ...")
				dns.DropHosts()
				log.Debug().Msg("Cleaning DNS configuration ...")
				dns.Ins().RestoreNameServer()
			} else {
				log.Info().Msgf("Not %s user, DNS cleanup skipped", util.GetAdminUserName())
			}
		}
	}
	return nil
}

func cleanPidFiles() {
	files, _ := ioutil.ReadDir(util.KtHome)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pid") {
			component, pid := parseComponentAndPid(f.Name())
			if util.IsProcessExist(pid) {
				log.Debug().Msgf("Find kt %s instance with pid %d", component, pid)
			} else {
				log.Info().Msgf("Removing remnant pid file %s", f.Name())
				if err := os.Remove(fmt.Sprintf("%s/%s", util.KtHome, f.Name())); err != nil {
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
