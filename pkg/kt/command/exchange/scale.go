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

func ByScale(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ctx := context.Background()
	app, err := cli.Kubernetes().GetDeployment(ctx, deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	options.RuntimeOptions.Origin = deploymentName
	options.RuntimeOptions.Replicas = *app.Spec.Replicas

	shadowPodName := deploymentName + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))

	log.Info().Msgf("Creating exchange shadow %s in namespace %s", shadowPodName, options.Namespace)
	if err = general.CreateShadowAndInbound(ctx, cli.Kubernetes(), shadowPodName, getExchangeLabels(shadowPodName, app),
		getExchangeAnnotation(options), options); err != nil {
		return err
	}

	down := int32(0)
	if err = cli.Kubernetes().ScaleTo(ctx, deploymentName, options.Namespace, &down); err != nil {
		return err
	}

	return nil
}

func getExchangeAnnotation(options *options.DaemonOptions) map[string]string {
	return map[string]string{
		common.KtConfig: fmt.Sprintf("app=%s,replicas=%d",
			options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas),
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
