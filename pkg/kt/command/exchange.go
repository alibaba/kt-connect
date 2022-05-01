package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/exchange"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand(action ActionInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "exchange",
		Short: "Redirect all requests of specified kubernetes service to local",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("name of service to exchange is required")
			}
			return action.Exchange(args[0])
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl exchange <service-name> [command options]"))
	cmd.Long = cmd.Short

	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	cmd.Flags().StringVar(&opt.Get().ExchangeOptions.Expose, "expose", "", "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80")
	cmd.Flags().StringVar(&opt.Get().ExchangeOptions.Mode, "mode", util.ExchangeModeSelector, "Exchange method 'selector', 'scale' or 'ephemeral'(experimental)")
	cmd.Flags().IntVar(&opt.Get().ExchangeOptions.RecoverWaitTime, "recoverWaitTime", 120, "(scale method only) Seconds to wait for original deployment recover before turn off the shadow pod")
	_ = cmd.MarkFlagRequired("expose")
	return cmd
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(resourceName string) error {
	ch, err := general.SetupProcess(util.ComponentExchange)
	if err != nil {
		return err
	}

	if port := util.FindBrokenLocalPort(opt.Get().ExchangeOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	if opt.Get().ExchangeOptions.Mode == util.ExchangeModeScale {
		err = exchange.ByScale(resourceName)
	} else if opt.Get().ExchangeOptions.Mode == util.ExchangeModeEphemeral {
		err = exchange.ByEphemeralContainer(resourceName)
	} else if opt.Get().ExchangeOptions.Mode == util.ExchangeModeSelector {
		err = exchange.BySelector(resourceName)
	} else {
		err = fmt.Errorf("invalid exchange method '%s', supportted are %s, %s, %s", opt.Get().ExchangeOptions.Mode,
			util.ExchangeModeSelector, util.ExchangeModeScale, util.ExchangeModeEphemeral)
	}
	if err != nil {
		return err
	}
	resourceType, realName := toTypeAndName(resourceName)
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now all request to %s '%s' will be redirected to local", resourceType, realName)
	log.Info().Msg("---------------------------------------------------------------")

	// watch background process, clean the workspace and exit if background process occur exception
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return nil
}

func toTypeAndName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		return parts[0], parts[1]
	} else {
		return "service", parts[0]
	}
}
