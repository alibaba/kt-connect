package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"github.com/alibaba/kt-connect/pkg/kt/command/connect"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewConnectCommand return new connect command
func NewConnectCommand(action ActionInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "connect",
		Short: "Create a network tunnel to kubernetes cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := preCheck(); err != nil {
				return err
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Connect()
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl connect [command options]"))
	cmd.Long = cmd.Short

	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	cmd.Flags().StringVar(&opt.Get().ConnectOptions.Mode, "mode", util.ConnectModeTun2Socks, "Connect mode 'tun2socks' or 'sshuttle'")
	cmd.Flags().StringVar(&opt.Get().ConnectOptions.DnsMode, "dnsMode", util.DnsModeLocalDns, "Specify how to resolve service domains, can be 'localDNS', 'podDNS', 'hosts' or 'hosts:<namespaces>', for multiple namespaces use ',' separation")
	cmd.Flags().BoolVar(&opt.Get().ConnectOptions.SharedShadow, "shareShadow", false, "Use shared shadow pod")
	cmd.Flags().StringVar(&opt.Get().ConnectOptions.ClusterDomain, "clusterDomain", "cluster.local", "The cluster domain provided to kubernetes api-server")
	cmd.Flags().BoolVar(&opt.Get().ConnectOptions.DisablePodIp, "disablePodIp", false, "Disable access to pod IP address")
	cmd.Flags().BoolVar(&opt.Get().ConnectOptions.SkipCleanup, "skipCleanup", false, "Do not auto cleanup residual resources in cluster")
	cmd.Flags().StringVar(&opt.Get().ConnectOptions.IncludeIps, "includeIps", "", "Specify extra IP ranges which should be route to cluster, e.g. '172.2.0.0/16', use ',' separated")
	cmd.Flags().StringVar(&opt.Get().ConnectOptions.ExcludeIps, "excludeIps", "", "Do not route specified IPs to cluster, e.g. '192.168.64.2' or '192.168.64.0/24', use ',' separated")
	cmd.Flags().BoolVar(&opt.Get().ConnectOptions.DisableTunDevice, "disableTunDevice", false, "(tun2socks mode only) Create socks5 proxy without tun device")
	cmd.Flags().BoolVar(&opt.Get().ConnectOptions.DisableTunRoute, "disableTunRoute", false, "(tun2socks mode only) Do not auto setup tun device route")
	cmd.Flags().IntVar(&opt.Get().ConnectOptions.SocksPort, "proxyPort", 2223, "(tun2socks mode only) Specify the local port which socks5 proxy should use")
	cmd.Flags().Int64Var(&opt.Get().ConnectOptions.DnsCacheTtl, "dnsCacheTtl", 60, "(local dns mode only) DNS cache refresh interval in seconds")
	return cmd
}

// Connect setup vpn to kubernetes cluster
func (action *Action) Connect() error {
	ch, err := general.SetupProcess(util.ComponentConnect)
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
	s := <-ch
	log.Info().Msgf("Terminal signal is %s", s)
	return nil
}

func preCheck() error {
	if err := checkPermissionAndOptions(); err != nil {
		return err
	}
	if pid := util.GetDaemonRunning(util.ComponentConnect); pid > 0 {
		return fmt.Errorf("another connect process already running at %d, exiting", pid)
	}
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
