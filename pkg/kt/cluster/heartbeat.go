package cluster

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	"strconv"
	"time"
)

// setup a heartbeat watcher
func setupHeartBeat(client v1.DeploymentInterface, name string) {
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			refreshHeartBeat(client, name)
		}
	}()
}

func refreshHeartBeat(client v1.DeploymentInterface, name string) {
	now := time.Now()
	log.Info().Msgf("heart beat ticked at %s", now.Format(common.YyyyMmDdHhMmSs))
	value := fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		common.KTLastHeartBeat, strconv.FormatInt(now.Unix(), 10))
	_, err := client.Patch(name, types.JSONPatchType, []byte(value))
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}
