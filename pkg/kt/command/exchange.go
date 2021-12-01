package command

import (
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/exchange"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
)

// NewExchangeCommand return new exchange command
func NewExchangeCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: general.ExchangeActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
				return err
			}
			resourceToExchange := c.Args().First()
			expose := options.ExchangeOptions.Expose

			if len(resourceToExchange) == 0 {
				return errors.New("name of resource to exchange is required")
			}
			if len(expose) == 0 {
				return errors.New("--expose is required")
			}

			return action.Exchange(resourceToExchange, cli, options)
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

	method := options.ExchangeOptions.Method
	if method == common.ExchangeMethodScale {
		err = exchange.ExchangeByScale(resourceName, cli, options)
	} else if method == common.ExchangeMethodEphemeral {
		err = exchange.ExchangeByEphemeralContainer(resourceName, cli, options)
	} else {
		err = fmt.Errorf("invalid exchange method \"%s\"", method)
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
	return nil
}
