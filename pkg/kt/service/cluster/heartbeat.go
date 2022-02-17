package cluster

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

const ResourceHeartBeatIntervalMinus = 2
const portForwardHeartBeatIntervalSec = 60

// SetupHeartBeat setup heartbeat watcher
func SetupHeartBeat(name, namespace string, updater func(string, string)) {
	ticker := time.NewTicker(time.Minute *ResourceHeartBeatIntervalMinus - util.RandomSeconds(0, 10))
	go func() {
		for range ticker.C {
			updater(name, namespace)
		}
	}()
}

// SetupPortForwardHeartBeat setup heartbeat watcher for port forward
func SetupPortForwardHeartBeat(port int) {
	ticker := time.NewTicker(time.Second *portForwardHeartBeatIntervalSec - util.RandomSeconds(0, 5))
	go func() {
		for range ticker.C {
			if conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port)); err != nil {
				log.Warn().Err(err).Msgf("Heartbeat port forward %d ticked failed: %s", port, err)
			} else {
				log.Debug().Msgf("Heartbeat port forward %d ticked at %s", port, formattedTime())
				_ = conn.Close()
			}
		}
	}()
}

func formattedTime() string {
	return time.Now().Format(common.YyyyMmDdHhMmSs)
}

func resourceHeartbeatPatch() string {
	return fmt.Sprintf("[ { \"op\" : \"replace\" , \"path\" : \"/metadata/annotations/%s\" , \"value\" : \"%s\" } ]",
		util.KtLastHeartBeat, util.GetTimestamp())
}
