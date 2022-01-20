package exchange

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"strings"
)

func BySelector(ctx context.Context, k cluster.KubernetesInterface, resourceName string, opts *options.DaemonOptions) error {
	// Get service to exchange
	svc, err := general.GetServiceByResourceName(ctx, k, resourceName, opts.Namespace)
	if err != nil {
		return err
	}

	// Lock service to avoid conflict
	if err = general.LockService(ctx, k, svc.Name, opts.Namespace, 0); err != nil {
		return err
	}
	defer general.UnlockService(ctx, k, svc.Name, opts.Namespace)

	// Create shadow pod
	shadowName := svc.Name + common.ExchangePodInfix + strings.ToLower(util.RandomString(5))
	shadowLabels := map[string]string{
		common.KtRole: common.RoleExchangeShadow,
		common.KtName: shadowName,
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