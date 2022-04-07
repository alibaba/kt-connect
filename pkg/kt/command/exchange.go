package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/exchange"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand(action ActionInterface, ch chan os.Signal) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "exchange",
		Long: "Redirect all requests of specified kubernetes service to local",
		Short: "ktctl exchange <service-name> [command options]",
		Run: func(cmd *cobra.Command, args []string) {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				log.Error().Msgf("%s", err)
			} else if len(args) == 0 {
				log.Error().Msgf("name of service to exchange is required")
			} else if len(opt.Get().ExchangeOptions.Expose) == 0 {
				log.Error().Msgf("--expose is required")
			} else if err2 := action.Exchange(args[0], ch); err2 != nil {
				log.Error().Msgf("%s", err2)
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&opt.Get().ExchangeOptions.Expose, "expose", util.ExchangeModeSelector, "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80")
	cmd.Flags().StringVar(&opt.Get().ExchangeOptions.Mode, "mode", util.MeshModeAuto, "Exchange method 'selector', 'scale' or 'ephemeral'(experimental)")
	cmd.Flags().IntVar(&opt.Get().ExchangeOptions.RecoverWaitTime, "recoverWaitTime", 120, "(scale method only) Seconds to wait for original deployment recover before turn off the shadow pod")
	_ = cmd.MarkFlagRequired("expose")
	return cmd
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(resourceName string, ch chan os.Signal) error {
	err := general.SetupProcess(util.ComponentExchange, ch)
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
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		ch <-os.Interrupt
	}()
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
