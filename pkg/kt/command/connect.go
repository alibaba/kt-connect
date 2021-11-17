package command

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"net"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/cilium/ipam/service/allocator"
	"github.com/cilium/ipam/service/ipallocator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
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
			if err := completeOptions(options); err != nil {
				return err
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

	if err = connectToCluster(context.TODO(), cli, options); err != nil {
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
	log.Info().Msgf("Terminal signal is %s", s)
	return nil
}

func completeOptions(options *options.DaemonOptions) error {
	if options.ConnectOptions.Method == common.ConnectMethodTun {
		srcIP, destIP, err := allocateTunIP(options.ConnectOptions.TunCidr)
		if err != nil {
			return err
		}
		options.ConnectOptions.SourceIP = srcIP
		options.ConnectOptions.DestIP = destIP
	}

	return nil
}

func connectToCluster(ctx context.Context, cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return
	}

	if options.ConnectOptions.Method == common.ConnectMethodSocks ||
		options.ConnectOptions.Method == common.ConnectMethodSocks5 {
		setupDump2Host(options, kubernetes)
	}
	if options.ConnectOptions.Method == common.ConnectMethodVpn {
		checkSshuttleInstalled(cli.Exec().Sshuttle())
	} else if options.ConnectOptions.UseGlobalProxy {
		setupGlobalProxy(options)
	}

	endPointIP, podName, credential, err := getOrCreateShadow(options, err, kubernetes)
	if err != nil {
		return
	}

	cidrs, err := kubernetes.ClusterCidrs(ctx, options.Namespace, options.ConnectOptions)
	if err != nil {
		return
	}

	return cli.Shadow().Outbound(podName, endPointIP, credential, cidrs, cli.Exec())
}

func checkSshuttleInstalled(cli sshuttle.CliInterface) {
	if !exec.CanRun(cli.Version()) {
		err := exec.RunAndWait(cli.Install(), "install_sshuttle")
		if err != nil {
			log.Error().Err(err).Msgf("Failed find or install sshuttle")
		}
	}
}

func setupGlobalProxy(options *options.DaemonOptions) {
	var err error
	if options.ConnectOptions.Method == common.ConnectMethodSocks {
		err = registry.SetGlobalProxy(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to setup global connect proxy")
		}
	}
	err = registry.SetHttpProxyEnvironmentVariable(options.ConnectOptions.Method, options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to setup global http proxy")
	}
}

func getOrCreateShadow(options *options.DaemonOptions, err error, kubernetes cluster.KubernetesInterface) (string, string, *util.SSHCredential, error) {
	shadowPodName := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.ShareShadow {
		shadowPodName = fmt.Sprintf("kt-connect-daemon-shared")
	}
	annotations := make(map[string]string)

	endPointIP, podName, sshConfigMapName, credential, err := cluster.GetOrCreateShadow(context.TODO(), kubernetes,
		shadowPodName, options, getLabels(shadowPodName, options), annotations, getEnvs(options))
	if err != nil {
		return "", "", nil, err
	}

	// record shadow name will clean up terminal
	options.RuntimeOptions.Shadow = shadowPodName
	options.RuntimeOptions.SSHCM = sshConfigMapName

	return endPointIP, podName, credential, nil
}

func setupDump2Host(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	namespacesToDump := []string{options.Namespace}
	if options.ConnectOptions.Dump2HostsNamespaces != "" {
		for _, ns := range strings.Split(options.ConnectOptions.Dump2HostsNamespaces, ",") {
			namespacesToDump = append(namespacesToDump, ns)
		}
	}
	hosts := map[string]string{}
	for _, namespace := range namespacesToDump {
		log.Debug().Msgf("Search service in %s namespace ...", namespace)
		singleHosts := kubernetes.GetServiceHosts(context.TODO(), namespace)
		for svc, ip := range singleHosts {
			if ip == "" || ip == "None" {
				continue
			}
			log.Debug().Msgf("Service found: %s.%s %s", svc, namespace, ip)
			if namespace == options.Namespace {
				hosts[svc] = ip
			}
			hosts[svc+"."+namespace] = ip
			hosts[svc+"."+namespace+"."+options.ConnectOptions.ClusterDomain] = ip
		}
	}
	options.RuntimeOptions.Dump2Host = util.DumpHosts(hosts)
}

func getEnvs(options *options.DaemonOptions) map[string]string {
	envs := make(map[string]string)
	localDomains := util.GetLocalDomains()
	if localDomains != "" {
		log.Debug().Msgf("Found local domains: %s", localDomains)
		envs[common.EnvVarLocalDomains] = localDomains
	}
	if options.ConnectOptions.Method == common.ConnectMethodTun {
		envs[common.ClientTunIP] = options.ConnectOptions.SourceIP
		envs[common.ServerTunIP] = options.ConnectOptions.DestIP
		envs[common.TunMaskLength] = util.ExtractNetMaskFromCidr(options.ConnectOptions.TunCidr)
	}
	return envs
}

func getLabels(workload string, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KtComponent: common.ComponentConnect,
		common.KtName:      workload,
	}
	splits := strings.Split(workload, "-")
	labels[common.KtVersion] = splits[len(splits)-1]
	return labels
}

func allocateTunIP(cidr string) (srcIP, destIP string, err error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	rge, err := ipallocator.NewAllocatorCIDRRange(ipnet, func(max int, rangeSpec string) (allocator.Interface, error) {
		return allocator.NewContiguousAllocationMap(max, rangeSpec), nil
	})
	if err == nil {
		ip1, _ := rge.AllocateNext()
		ip2, _ := rge.AllocateNext()
		if ip1 != nil && ip2 != nil {
			return ip1.String(), ip2.String(), nil
		}
	}
	return "", "", err
}
