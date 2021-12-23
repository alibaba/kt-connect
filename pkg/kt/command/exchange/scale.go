package exchange

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	"strings"
)

func ByScale(deploymentName string, cli kt.CliInterface, opts *options.DaemonOptions) error {
	ctx := context.Background()
	app, err := cli.Kubernetes().GetDeployment(ctx, deploymentName, opts.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	opts.RuntimeOptions.Origin = deploymentName
	opts.RuntimeOptions.Replicas = *app.Spec.Replicas

	shadowPodName := deploymentName + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))

	log.Info().Msgf("Creating exchange shadow %s in namespace %s", shadowPodName, opts.Namespace)
	if err = general.CreateShadowAndInbound(ctx, cli.Kubernetes(), shadowPodName, opts.ExchangeOptions.Expose,
		getExchangeLabels(shadowPodName, app), getExchangeAnnotation(opts), opts); err != nil {
		return err
	}

	down := int32(0)
	if err = cli.Kubernetes().ScaleTo(ctx, deploymentName, opts.Namespace, &down); err != nil {
		return err
	}

	return nil
}

func getExchangeAnnotation(opts *options.DaemonOptions) map[string]string {
	return map[string]string{
		common.KtConfig: fmt.Sprintf("app=%s,replicas=%d",
			opts.RuntimeOptions.Origin, opts.RuntimeOptions.Replicas),
	}
}

func getExchangeLabels(workload string, origin *appV1.Deployment) map[string]string {
	labels := map[string]string{
		common.KtRole:    common.RoleExchangeShadow,
		common.KtName:    workload,
	}
	if origin != nil {
		for k, v := range origin.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	return labels
}
