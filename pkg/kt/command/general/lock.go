package general

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	"time"
)

// LockTimeout 3 minutes
const LockTimeout = 3 * 60

func LockService(serviceName, namespace string, times int) (*coreV1.Service, error) {
	if times > 10 {
		return nil, fmt.Errorf("failed to obtain kt lock of service %s, please try again later", serviceName)
	}
	svc, err := cluster.Ins().GetService(serviceName, namespace)
	if err != nil {
		return nil, err
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	if lock, exists := svc.Annotations[util.KtLock]; exists && util.GetTime() - util.ParseTimestamp(lock) < LockTimeout {
		log.Info().Msgf("Another user is occupying service %s, waiting for lock ...", serviceName)
		time.Sleep(3 * time.Second)
		return LockService(serviceName, namespace, times + 1)
	} else {
		svc.Annotations[util.KtLock] = util.GetTimestamp()
		if svc, err = cluster.Ins().UpdateService(svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock service %s", serviceName)
			return LockService(serviceName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Service %s locked", serviceName)
	return svc, nil
}

func UnlockService(serviceName, namespace string) {
	svc, err := cluster.Ins().GetService(serviceName, namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get service %s for unlock", serviceName)
		return
	}
	if _, exists := svc.Annotations[util.KtLock]; exists {
		delete(svc.Annotations, util.KtLock)
		if _, err = cluster.Ins().UpdateService(svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to unlock service %s", serviceName)
		} else {
			log.Info().Msgf("Service %s unlocked", serviceName)
		}
	} else {
		log.Info().Msgf("Service %s doesn't have lock", serviceName)
	}
}
