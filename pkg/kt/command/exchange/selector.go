package exchange

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func BySelector(resourceName string) error {
	// Get service to exchange
	svc, err := general.GetServiceByResourceName(resourceName, opt.Get().Global.Namespace)
	if err != nil {
		return err
	}
	if port := util.FindInvalidRemotePort(opt.Get().Exchange.Expose, general.GetTargetPorts(svc)); port != "" {
		return fmt.Errorf("target port %s not exists in service %s", port, svc.Name)
	}

	// Lock service to avoid conflict, must be first step
	svc, err = general.LockService(svc.Name, opt.Get().Global.Namespace, 0);
	if err != nil {
		return err
	}
	defer general.UnlockService(svc.Name, opt.Get().Global.Namespace)

	if svc.Annotations != nil && svc.Annotations[util.KtSelector] != "" {
		if svc.Spec.Selector[util.KtRole] == util.RoleExchangeShadow {
			return fmt.Errorf("service '%s' is already exchanging by another user%s, cannot apply exchange",
				svc.Name, general.GetOccupiedUser(svc.Spec.Selector))
		} else if svc.Spec.Selector[util.KtRole] == util.RoleRouter {
			return fmt.Errorf("another user is meshing service '%s', cannot apply exchange", svc.Name)
		} else {
			log.Warn().Msgf("Service '%s' has %s annotation, but either selecting shadow or router pod", svc.Name, util.KtSelector)
			return fmt.Errorf("service '%s' in invalid status, please manually remove %s annotation before exchange", svc.Name, util.KtSelector)
		}
	}

	// Create shadow pod
	shadowName := svc.Name + util.ExchangePodInfix + strings.ToLower(util.RandomString(5))
	shadowLabels := map[string]string{
		util.KtRole:   util.RoleExchangeShadow,
		util.KtTarget: util.RandomString(20),
	}
	annotation := map[string]string{
		util.KtConfig: fmt.Sprintf("service=%s", svc.Name),
	}
	if err = general.CreateShadowAndInbound(shadowName, opt.Get().Exchange.Expose,
		shadowLabels, annotation, general.GetTargetPorts(svc)); err != nil {
		return err
	}

	// Let target service select shadow pod
	opt.Store.Origin = svc.Name
	if err = general.UpdateServiceSelector(svc.Name, opt.Get().Global.Namespace, shadowLabels); err != nil {
		return err
	}

	return nil
}
