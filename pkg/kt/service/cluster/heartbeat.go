package cluster

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
	"time"
)

const ResourceHeartBeatIntervalMinus = 2
const portForwardHeartBeatIntervalSec = 60

var TimeDifference int64 = 0

// SetupTimeDifference get time difference between cluster and local
func SetupTimeDifference() error {
	rectifierPodName := fmt.Sprintf("kt-rectifier-%s", strings.ToLower(util.RandomString(5)))
	_, err := Ins().CreateRectifierPod(rectifierPodName)
	if err != nil {
		return err
	}
	stdout, stderr, err := Ins().ExecInPod(util.DefaultContainer, rectifierPodName, opt.Get().Namespace, "date", "+%s")
	if err != nil {
		return err
	}
	go func() {
		if err2 := Ins().RemovePod(rectifierPodName, opt.Get().Namespace); err2 != nil {
			log.Debug().Err(err).Msgf("Failed to remove pod %s", rectifierPodName)
		}
	}()
	remoteTime, err := strconv.ParseInt(stdout, 10, 0)
	if err != nil {
		log.Warn().Msgf("Invalid cluster time: '%s' %s", stdout, stderr)
		return err
	}
	TimeDifference = remoteTime - time.Now().Unix()
	if TimeDifference >= -1 && TimeDifference <= 1 {
		log.Info().Msgf("No time difference")
	} else {
		log.Info().Msgf("Time difference is %d", TimeDifference)
	}
	return nil
}

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
