package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/connect"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
)

// NewConnectCommand return new connect command
func NewConnectCommand(action ActionInterface, ch chan os.Signal) urfave.Command {
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
			return action.Connect(ch)
		},
	}
}

// Connect setup vpn to kubernetes cluster
func (action *Action) Connect(ch chan os.Signal) error {
	if err := checkPermissionAndOptions(); err != nil {
		return err
	}
	if pid := util.GetDaemonRunning(util.ComponentConnect); pid > 0 {
		return fmt.Errorf("another connect process already running at %d, exiting", pid)
	}

	err := general.SetupProcess(util.ComponentConnect, ch)
	if err != nil {
		return err
	}

	if !opt.Get().ConnectOptions.SkipCleanup {
		go silenceCleanup()
	}

	if opt.Get().ConnectOptions.Mode == util.ConnectModeTun2Socks {
		err = connect.ByTun2Socks()
	} else if opt.Get().ConnectOptions.Mode == util.ConnectModeShuttle {
		err = connect.BySshuttle()
	} else {
		err = fmt.Errorf("invalid connect mode: '%s', supportted mode are %s, %s", opt.Get().ConnectOptions.Mode,
			util.ConnectModeTun2Socks, util.ConnectModeShuttle)
	}
	if err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" All looks good, now you can access to resources in the kubernetes cluster")
	log.Info().Msg("---------------------------------------------------------------")

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		ch <- os.Interrupt
	}()
	s := <-ch
	log.Info().Msgf("Terminal signal is %s", s)
	return nil
}

func silenceCleanup() {
	if r, err := clean.CheckClusterResources(); err == nil {
		for _, name := range r.PodsToDelete {
			_ = cluster.Ins().RemovePod(name, opt.Get().Namespace)
		}
		for _, name := range r.ConfigMapsToDelete {
			_ = cluster.Ins().RemoveConfigMap(name, opt.Get().Namespace)
		}
		for _, name := range r.DeploymentsToDelete {
			_ = cluster.Ins().RemoveDeployment(name, opt.Get().Namespace)
		}
		for _, name := range r.ServicesToDelete {
			_ = cluster.Ins().RemoveService(name, opt.Get().Namespace)
		}
	}
}

func checkPermissionAndOptions() error {
	if !util.IsRunAsAdmin() {
		if util.IsWindows() {
			return fmt.Errorf("permission declined, please re-run connect command as Administrator")
		}
		return fmt.Errorf("permission declined, please re-run connect command with 'sudo'")
	}
	if opt.Get().ConnectOptions.Mode == util.ConnectModeTun2Socks && opt.Get().ConnectOptions.DnsMode == util.DnsModePodDns {
		return fmt.Errorf("dns mode '%s' is not available for connect mode '%s'", util.DnsModePodDns, util.ConnectModeTun2Socks)
	}
	return nil
}
