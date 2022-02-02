package exchange

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func BySelector(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// Get service to exchange
	svc, err := general.GetServiceByResourceName(ctx, k, resourceName, opts.Namespace)
	if err != nil {
		return err
	}

	// Lock service to avoid conflict, must be first step
	svc, err = general.LockService(ctx, k, svc.Name, opts.Namespace, 0);
	if err != nil {
		return err
	}
	defer general.UnlockService(ctx, k, svc.Name, opts.Namespace)

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
	if err = general.CreateShadowAndInbound(ctx, k, shadowName, opts.ExchangeOptions.Expose, shadowLabels, map[string]string{}, opts); err != nil {
		return err
	}

	// Let target service select shadow pod
	opts.RuntimeOptions.Origin = svc.Name
	if err = general.UpdateServiceSelector(ctx, k, svc.Name, opts.Namespace, shadowLabels); err != nil {
		return err
	}

	return nil
}