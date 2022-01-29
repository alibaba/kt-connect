package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/exchange"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
	"time"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "exchange",
		Usage: "redirect all requests of specified kubernetes service to local",
		UsageText: "ktctl exchange <service-name> [command options]",
		Flags: general.ExchangeActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
				return err
			}

			if len(c.Args()) == 0 {
				return errors.New("name of resource to exchange is required")
			}
			if len(options.ExchangeOptions.Expose) == 0 {
				return errors.New("--expose is required")
			}

			return action.Exchange(c.Args().First(), cli, options)
		},
	}
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(resourceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch, err := general.SetupProcess(cli, options, common.ComponentExchange)
	if err != nil {
		return err
	}

	if port := util.FindBrokenPort(options.ExchangeOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	method := options.ExchangeOptions.Mode
	if method == common.ExchangeModeScale {
		err = exchange.ByScale(resourceName, cli, options)
	} else if method == common.ExchangeModeEphemeral {
		err = exchange.ByEphemeralContainer(resourceName, cli, options)
	} else if method == common.ExchangeModeSelector {
		err = exchange.BySelector(context.TODO(), cli.Kubernetes(), resourceName, options)
	} else {
		err = fmt.Errorf("invalid exchange method '%s'", method)
	}
	if err != nil {
		return err
	}

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	// when process interrupt by signal, wait a while for resource clean up
	time.Sleep(1 * time.Second)
	return nil
}
