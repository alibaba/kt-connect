package preview

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

// Expose create a new service in cluster
func Expose(ctx context.Context, serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	version := strings.ToLower(util.RandomString(5))
	shadowPodName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RolePreviewShadow,
		common.KtName:    shadowPodName,
	}
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", serviceName),
	}

	return exposeLocalService(ctx, serviceName, shadowPodName, labels, annotations, options, cli.Kubernetes())
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(ctx context.Context, serviceName, shadowPodName string, labels, annotations map[string]string,
	options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) error {

	envs := make(map[string]string)
	_, podName, credential, err := cluster.GetOrCreateShadow(ctx, kubernetes, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}
	log.Info().Msgf("Created shadow pod %s", podName)

	portPairs := strings.Split(options.PreviewOptions.Expose, ",")
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
			Namespace: options.Namespace,
			Labels: map[string]string{},
			Annotations: map[string]string{},
		},
		External: options.PreviewOptions.External,
		Ports: ports,
		Selectors: labels,
	}); err != nil {
		return err
	}
	options.RuntimeOptions.Service = serviceName

	if _, err = tunnel.ForwardPodToLocal(options.PreviewOptions.Expose, podName, credential.PrivateKeyPath, options); err != nil {
		return err
	}

	log.Info().Msgf("Forward remote %s:%v -> 127.0.0.1:%v", podName, options.PreviewOptions.Expose, options.PreviewOptions.Expose)
	return nil
}
