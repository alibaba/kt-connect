package exchange

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func BySelector(ctx context.Context, resourceName string) error {
	// Get service to exchange
	svc, err := general.GetServiceByResourceName(ctx, resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}

	// Lock service to avoid conflict, must be first step
	svc, err = general.LockService(ctx, svc.Name, opt.Get().Namespace, 0);
	if err != nil {
		return err
	}
	defer general.UnlockService(ctx, svc.Name, opt.Get().Namespace)

	if svc.Annotations != nil && svc.Annotations[common.KtSelector] != "" {
		if svc.Spec.Selector[common.KtRole] == common.RoleExchangeShadow {
			return fmt.Errorf("service '%s' is already exchanging by another user, cannot apply exchange", svc.Name)
		} else if svc.Spec.Selector[common.KtRole] == common.RoleRouter {
			return fmt.Errorf("another user is meshing service '%s', cannot apply exchange", svc.Name)
		} else {
			log.Warn().Msgf("Service '%s' has %s annotation, but either selecting shadow or router pod", svc.Name, common.KtSelector)
			return fmt.Errorf("service '%s' in invalid status, please manually remove %s annotation before exchange", svc.Name, common.KtSelector)
		}
	}

	// Create shadow pod
	shadowName := svc.Name + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))
	shadowLabels := map[string]string{
		common.KtRole: common.RoleExchangeShadow,
		common.KtTarget: util.RandomString(20),
	}
	if err = general.CreateShadowAndInbound(ctx, shadowName, opt.Get().ExchangeOptions.Expose, shadowLabels, map[string]string{}); err != nil {
		return err
	}

	// Let target service select shadow pod
	opt.Get().RuntimeOptions.Origin = svc.Name
	if err = general.UpdateServiceSelector(ctx, svc.Name, opt.Get().Namespace, shadowLabels); err != nil {
		return err
	}

	return nil
}