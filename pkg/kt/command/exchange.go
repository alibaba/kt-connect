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
	urfave "github.com/urfave/cli"
	"os"
	"strings"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand(action ActionInterface, ch chan os.Signal) urfave.Command {
	return urfave.Command{
		Name:  "exchange",
		Usage: "redirect all requests of specified kubernetes service to local",
		UsageText: "ktctl exchange <service-name> [command options]",
		Flags: general.ExchangeActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}

			if len(c.Args()) == 0 {
				return fmt.Errorf("name of resource to exchange is required")
			}
			if len(opt.Get().ExchangeOptions.Expose) == 0 {
				return fmt.Errorf("--expose is required")
			}

			return action.Exchange(c.Args().First(), ch)
		},
	}
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
