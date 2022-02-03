package preview

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

// Expose create a new service in cluster
func Expose(ctx context.Context, serviceName string, cli kt.CliInterface) error {
	version := strings.ToLower(util.RandomString(5))
	shadowPodName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RolePreviewShadow,
		common.KtTarget:  util.RandomString(20),
	}
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", serviceName),
	}

	return exposeLocalService(ctx, serviceName, shadowPodName, labels, annotations, cli.Kubernetes())
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(ctx context.Context, serviceName, shadowPodName string, labels, annotations map[string]string,
	kubernetes cluster.KubernetesInterface) error {

	envs := make(map[string]string)
	_, podName, credential, err := cluster.GetOrCreateShadow(ctx, kubernetes, shadowPodName, labels, annotations, envs)
	if err != nil {
		return err
	}
	log.Info().Msgf("Created shadow pod %s", podName)

	portPairs := strings.Split(opt.Get().PreviewOptions.Expose, ",")
	ports := make(map[int]int)
	for _, exposePort := range portPairs {
		localPort, remotePort, err2 := util.ParsePortMapping(exposePort)
		if err2 != nil {
			return err
		}
		ports[localPort] = remotePort
	}
	if _, err = kubernetes.CreateService(ctx, &cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name: serviceName,
			Namespace: opt.Get().Namespace,
			Labels: map[string]string{},
			Annotations: map[string]string{},
		},
		External: opt.Get().PreviewOptions.External,
		Ports: ports,
		Selectors: labels,
	}); err != nil {
		return err
	}
	opt.Get().RuntimeOptions.Service = serviceName

	if _, err = transmission.ForwardPodToLocal(opt.Get().PreviewOptions.Expose, podName, credential.PrivateKeyPath); err != nil {
		return err
	}

	log.Info().Msgf("Forward remote %s:%v -> 127.0.0.1:%v", podName, opt.Get().PreviewOptions.Expose, opt.Get().PreviewOptions.Expose)
	return nil
}
