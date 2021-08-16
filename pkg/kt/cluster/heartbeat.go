package cluster

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
	appV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

// setupDeploymentHeartBeat setup heartbeat watcher for deployment
func setupDeploymentHeartBeat(client appV1.DeploymentInterface, name string) {
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			refreshDeploymentHeartBeat(client, name)
		}
	}()
}

func refreshDeploymentHeartBeat(client appV1.DeploymentInterface, name string) {
	log.Info().Msgf("Heart beat ticked at %s", time.Now().Format(common.YyyyMmDdHhMmSs))
	value := fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		common.KTLastHeartBeat, util.GetTimestamp())
	_, err := client.Patch(name, types.JSONPatchType, []byte(value))
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}

// setupServiceHeartBeat setup heartbeat watcher for service
func setupServiceHeartBeat(client v1.ServiceInterface, name string) {
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			refreshServiceHeartBeat(client, name)
		}
	}()
}

func refreshServiceHeartBeat(client v1.ServiceInterface, name string) {
	log.Info().Msgf("Heart beat ticked at %s", time.Now().Format(common.YyyyMmDdHhMmSs))
	value := fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		common.KTLastHeartBeat, util.GetTimestamp())
	_, err := client.Patch(name, types.JSONPatchType, []byte(value))
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}

// setupConfigMapHeartBeat setup heartbeat watcher for config map
func setupConfigMapHeartBeat(client v1.ConfigMapInterface, name string) {
	ticker := time.NewTicker(time.Minute * 10)
	go func() {
		for _ = range ticker.C {
			refreshConfigMapHeartBeat(client, name)
		}
	}()
}

func refreshConfigMapHeartBeat(client v1.ConfigMapInterface, name string) {
	log.Info().Msgf("Heart beat ticked at %s", time.Now().Format(common.YyyyMmDdHhMmSs))
	value := fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		common.KTLastHeartBeat, util.GetTimestamp())
	_, err := client.Patch(name, types.JSONPatchType, []byte(value))
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
}
