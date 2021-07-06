package command

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

// setup a heartbeat watcher
func setupHeartBeat(cli kt.CliInterface, options *options.DaemonOptions, podName string) {
	refreshHeartBeat(cli, options.Namespace, podName)
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			refreshHeartBeat(cli, options.Namespace, podName)
		}
	}()
}

func refreshHeartBeat(cli kt.CliInterface, namespace, podName string) {
	now := time.Now()
	log.Info().Msgf("heart beat ticked at %s", now.Format(common.YyyyMmDdHhMmSs))
	cli.Exec().Kubectl().UpdateAnnotate(namespace, podName, common.KTLastHeartBeat, strconv.FormatInt(now.Unix(), 10))
}
