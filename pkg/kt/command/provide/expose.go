package provide

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
	"strconv"
	"strings"
)

// Expose create a new service in cluster
func Expose(ctx context.Context, serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	version := strings.ToLower(util.RandomString(5))
	shadowPodName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RoleProvideShadow,
		common.KtName:    shadowPodName,
	}
	annotations := map[string]string{
		common.KtConfig: fmt.Sprintf("service=%s", serviceName),
	}

	return exposeLocalService(ctx, serviceName, shadowPodName, labels, annotations, options, kubernetes, cli)
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(ctx context.Context, serviceName, shadowPodName string, labels, annotations map[string]string,
	options *options.DaemonOptions, kubernetes cluster.KubernetesInterface, cli kt.CliInterface) (err error) {

	envs := make(map[string]string)
	_, podName, _, err := cluster.GetOrCreateShadow(ctx, kubernetes, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}
	log.Info().Msgf("Created shadow pod %s", podName)

	log.Info().Msgf("Expose deployment %s to service %s:%v", shadowPodName, serviceName, options.ProvideOptions.Expose)
	ports := map[int]int {options.ProvideOptions.Expose: options.ProvideOptions.Expose}
	if _, err = kubernetes.CreateService(ctx, &cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name: serviceName,
			Namespace: options.Namespace,
			Labels: map[string]string{},
			Annotations: map[string]string{},
		},
		External: options.ProvideOptions.External,
		Ports: ports,
		Selectors: labels,
	}); err != nil {
		return err
	}
	options.RuntimeOptions.Service = serviceName
	options.RuntimeOptions.Shadow = shadowPodName

	if _, err = tunnel.ForwardPodToLocal(strconv.Itoa(options.ProvideOptions.Expose), podName, options); err != nil {
		return err
	}

	log.Info().Msgf("Forward remote %s:%v -> 127.0.0.1:%v", podName, options.ProvideOptions.Expose, options.ProvideOptions.Expose)
	return nil
}
