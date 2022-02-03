package exchange

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	"strings"
)

func ByScale(resourceName string) error {
	ctx := context.Background()
	app, err := general.GetDeploymentByResourceName(ctx, resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	opt.Get().RuntimeOptions.Origin = resourceName
	opt.Get().RuntimeOptions.Replicas = *app.Spec.Replicas

	shadowPodName := resourceName + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))

	log.Info().Msgf("Creating exchange shadow %s in namespace %s", shadowPodName, opt.Get().Namespace)
	if err = general.CreateShadowAndInbound(ctx, shadowPodName, opt.Get().ExchangeOptions.Expose,
		getExchangeLabels(app), getExchangeAnnotation()); err != nil {
		return err
	}

	down := int32(0)
	if err = cluster.Ins().ScaleTo(ctx, resourceName, opt.Get().Namespace, &down); err != nil {
		return err
	}

	return nil
}

func getExchangeAnnotation() map[string]string {
	return map[string]string{
		common.KtConfig: fmt.Sprintf("app=%s,replicas=%d",
			opt.Get().RuntimeOptions.Origin, opt.Get().RuntimeOptions.Replicas),
	}
}

func getExchangeLabels(origin *appV1.Deployment) map[string]string {
	labels := map[string]string{
		common.KtRole: common.RoleExchangeShadow,
	}
	if origin != nil {
		for k, v := range origin.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	return labels
}
