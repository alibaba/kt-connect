package preview

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

// Expose create a new service in cluster
func Expose(serviceName string) error {
	version := strings.ToLower(util.RandomString(5))
	shadowPodName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		util.KtRole:    util.RolePreviewShadow,
		util.KtTarget:  util.RandomString(20),
	}
	annotations := map[string]string{
		util.KtConfig: fmt.Sprintf("service=%s", serviceName),
	}

	return exposeLocalService(serviceName, shadowPodName, labels, annotations)
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(serviceName, shadowPodName string, labels, annotations map[string]string) error {

	envs := make(map[string]string)
	_, podName, privateKeyPath, err := cluster.Ins().GetOrCreateShadow(shadowPodName, labels, annotations, envs,
		opt.Get().Preview.Expose, map[int]string{})
	if err != nil {
		return err
	}
	log.Info().Msgf("Created shadow pod %s", podName)

	portPairs := strings.Split(opt.Get().Preview.Expose, ",")
	ports := make(map[int]int)
	for _, exposePort := range portPairs {
		_, remotePort, err2 := util.ParsePortMapping(exposePort)
		if err2 != nil {
			return err
		}
		// service port to target port
		ports[remotePort] = remotePort
	}
	if _, err = cluster.Ins().CreateService(&cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        serviceName,
			Namespace:   opt.Get().Global.Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		External:  opt.Get().Preview.External,
		Ports:     ports,
		Selectors: labels,
	}); err != nil {
		return err
	}
	opt.Store.Service = serviceName

	if _, err = transmission.ForwardPodToLocal(opt.Get().Preview.Expose, podName, privateKeyPath); err != nil {
		return err
	}

	log.Info().Msgf("Forward remote %s:%v -> 127.0.0.1:%v", podName, opt.Get().Preview.Expose, opt.Get().Preview.Expose)
	return nil
}
