package util

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/types"
	appV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net"
	"time"
)

const ResourceHeartBeatIntervalMinus = 5
const portForwardHeartBeatIntervalSec = 30

// SetupDeploymentHeartBeat setup heartbeat watcher for deployment
func SetupDeploymentHeartBeat(client appV1.DeploymentInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat deployment %s ticked at %s", name, formattedTime())
			_, err := client.Patch(name, types.JSONPatchType, []byte(resourceHeartbeatPatch()))
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
		}
	}()
}

// SetupServiceHeartBeat setup heartbeat watcher for service
func SetupServiceHeartBeat(client v1.ServiceInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat service %s ticked at %s", name, formattedTime())
			_, err := client.Patch(name, types.JSONPatchType, []byte(resourceHeartbeatPatch()))
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
		}
	}()
}

// SetupConfigMapHeartBeat setup heartbeat watcher for config map
func SetupConfigMapHeartBeat(client v1.ConfigMapInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat configmap %s ticked at %s", name, formattedTime())
			_, err := client.Patch(name, types.JSONPatchType, []byte(resourceHeartbeatPatch()))
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
		}
	}()
}

// SetupPortForwardHeartBeat setup heartbeat watcher for port forward
func SetupPortForwardHeartBeat(port int) {
	ticker := time.NewTicker(time.Second * portForwardHeartBeatIntervalSec)
	go func() {
		for range ticker.C {
			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				log.Debug().Msgf("Heartbeat port forward %d ticked at %s", port, formattedTime())
				_ = conn.Close()
			} else {
				log.Debug().Msgf("Heartbeat port forward %d ticked failed %s", port, err)
			}
		}
	}()
}

func formattedTime() string {
	return time.Now().Format(common.YyyyMmDdHhMmSs)
}

func resourceHeartbeatPatch() string {
	return fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		common.KTLastHeartBeat, GetTimestamp())
}
