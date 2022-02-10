package general

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	"time"
)

// LockTimeout 5 minutes
const LockTimeout = 5 * 60

func LockService(ctx context.Context, serviceName, namespace string, times int) (*coreV1.Service, error) {
	if times > 10 {
		return nil, fmt.Errorf("failed to obtain kt lock of service %s, please try again later", serviceName)
	}
	svc, err := cluster.Ins().GetService(ctx, serviceName, namespace)
	if err != nil {
		return nil, err
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	if lock, ok := svc.Annotations[common.KtLock]; ok && time.Now().Unix() - util.ParseTimestamp(lock) < LockTimeout {
		log.Info().Msgf("Another user is occupying service %s, waiting for lock ...", serviceName)
		time.Sleep(3 * time.Second)
		return LockService(ctx, serviceName, namespace, times + 1)
	} else {
		svc.Annotations[common.KtLock] = util.GetTimestamp()
		if svc, err = cluster.Ins().UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to lock service %s", serviceName)
			return LockService(ctx, serviceName, namespace, times + 1)
		}
	}
	log.Info().Msgf("Service %s locked", serviceName)
	return svc, nil
}

func UnlockService(ctx context.Context, serviceName, namespace string) {
	svc, err := cluster.Ins().GetService(ctx, serviceName, namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get service %s for unlock", serviceName)
		return
	}
	if _, ok := svc.Annotations[common.KtLock]; ok {
		delete(svc.Annotations, common.KtLock)
		if _, err = cluster.Ins().UpdateService(ctx, svc); err != nil {
			log.Warn().Err(err).Msgf("Failed to unlock service %s", serviceName)
		} else {
			log.Info().Msgf("Service %s unlocked", serviceName)
		}
	} else {
		log.Info().Msgf("Service %s doesn't have lock", serviceName)
	}
}
