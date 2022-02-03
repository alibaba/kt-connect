package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command/connect"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
	"time"
)

// NewConnectCommand return new connect command
func NewConnectCommand(cli kt.CliInterface, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "connect",
		Usage: "create a network tunnel to kubernetes cluster",
		UsageText: "ktctl connect [command options]",
		Flags: general.ConnectActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}
			return action.Connect(cli)
		},
	}
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(cli kt.CliInterface) error {
	if err := checkOptions(); err != nil {
		return err
	}
	if pid := util.GetDaemonRunning(common.ComponentConnect); pid > 0 {
		return fmt.Errorf("another connect process already running at %d, exiting", pid)
	}

	ch, err := general.SetupProcess(cli, common.ComponentConnect)
	if err != nil {
		return err
	}

	if opt.Get().ConnectOptions.Mode == common.ConnectModeTun2Socks {
		err = connect.ByTun2Socks(cli)
	} else if opt.Get().ConnectOptions.Mode == common.ConnectModeShuttle {
		err = connect.BySshuttle(cli)
	} else {
		err = fmt.Errorf("invalid connect mode: '%s', supportted mode are %s, %s", opt.Get().ConnectOptions.Mode,
			common.ConnectModeTun2Socks, common.ConnectModeShuttle)
	}
	if err != nil {
		return err
	}
	log.Info().Msgf("All looks good, now you can access to resources in the kubernetes cluster")

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace(cli)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal signal is %s", s)
	// when process interrupt by signal, wait a while for resource clean up
	time.Sleep(1 * time.Second)
	return nil
}

func checkOptions() error {
	if opt.Get().ConnectOptions.Mode == common.ConnectModeTun2Socks && opt.Get().ConnectOptions.DnsMode == common.DnsModePodDns {
		return fmt.Errorf("dns mode '%s' is not available for connect mode '%s'", common.DnsModePodDns, common.ConnectModeTun2Socks)
	}
	return nil
}
