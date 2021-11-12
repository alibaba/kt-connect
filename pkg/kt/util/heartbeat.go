package util

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net"
	"time"
)

const ResourceHeartBeatIntervalMinus = 5
const portForwardHeartBeatIntervalSec = 30

// SetupPodHeartBeat setup heartbeat watcher for deployment
func SetupPodHeartBeat(ctx context.Context, client coreV1.PodInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat deployment %s ticked at %s", name, formattedTime())
			_, err := client.Patch(ctx, name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{})
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update pod heart beat")
				return
			}
		}
	}()
}

// SetupServiceHeartBeat setup heartbeat watcher for service
func SetupServiceHeartBeat(ctx context.Context, client coreV1.ServiceInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat service %s ticked at %s", name, formattedTime())
			_, err := client.Patch(ctx, name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{})
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update service heart beat")
				return
			}
		}
	}()
}

// SetupConfigMapHeartBeat setup heartbeat watcher for config map
func SetupConfigMapHeartBeat(ctx context.Context, client coreV1.ConfigMapInterface, name string) {
	ticker := time.NewTicker(time.Minute * ResourceHeartBeatIntervalMinus)
	go func() {
		for range ticker.C {
			log.Debug().Msgf("Heartbeat configmap %s ticked at %s", name, formattedTime())
			_, err := client.Patch(ctx, name, types.JSONPatchType, []byte(resourceHeartbeatPatch()), metav1.PatchOptions{})
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update config map heart beat")
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
				log.Debug().Msgf("Heartbeat port forward %d ticked failed: %s", port, err)
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
