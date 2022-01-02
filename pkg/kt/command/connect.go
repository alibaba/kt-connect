package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/connect"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
	"time"
)

// NewConnectCommand return new connect command
func NewConnectCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: general.ConnectActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
				return err
			}
			return action.Connect(cli, options)
		},
	}
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(cli kt.CliInterface, options *options.DaemonOptions) error {
	if pid := util.GetDaemonRunning(common.ComponentConnect); pid > 0 {
		return fmt.Errorf("another connect process already running at %d, exiting", pid)
	}

	ch, err := general.SetupProcess(cli, options, common.ComponentConnect)
	if err != nil {
		return err
	}

	if options.ConnectOptions.Method == common.ConnectMethodSocks5 {
		err = connect.BySocks5(cli, options)
	} else if options.ConnectOptions.Method == common.ConnectMethodVpn {
		err = connect.BySshuttle(cli, options)
	} else {
		err = fmt.Errorf("invalid connect mode: '%s'", options.ConnectOptions.Method)
	}
	if err != nil {
		return err
	}
	log.Info().Msgf("All looks good, now you can access to resources in the kubernetes cluster")

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal signal is %s", s)
	// when process interrupt by signal, wait a while for resource clean up
	time.Sleep(1 * time.Second)
	return nil
}

