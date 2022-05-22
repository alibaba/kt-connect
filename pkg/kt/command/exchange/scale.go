package exchange

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	"strings"
)

func ByScale(resourceName string) error {
	app, err := general.GetDeploymentByResourceName(resourceName, opt.Get().Global.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	opt.Store.Origin = app.Name
	opt.Store.Replicas = *app.Spec.Replicas

	shadowPodName := app.Name + util.ExchangePodInfix + strings.ToLower(util.RandomString(5))

	log.Info().Msgf("Creating exchange shadow %s in namespace %s", shadowPodName, opt.Get().Global.Namespace)
	if err = general.CreateShadowAndInbound(shadowPodName, opt.Get().Exchange.Expose,
		getExchangeLabels(app), getExchangeAnnotation(), map[int]string{}); err != nil {
		return err
	}

	down := int32(0)
	if err = cluster.Ins().ScaleTo(app.Name, opt.Get().Global.Namespace, &down); err != nil {
		return err
	}

	return nil
}

func getExchangeAnnotation() map[string]string {
	return map[string]string{
		util.KtConfig: fmt.Sprintf("app=%s,replicas=%d",
			opt.Store.Origin, opt.Store.Replicas),
	}
}

func getExchangeLabels(origin *appV1.Deployment) map[string]string {
	labels := map[string]string{
		util.KtRole: util.RoleExchangeShadow,
	}
	if origin != nil {
		for k, v := range origin.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	return labels
}
