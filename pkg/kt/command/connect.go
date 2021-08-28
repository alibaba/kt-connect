package command

import (
	"fmt"
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

// newConnectCommand return new connect command
func newConnectCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "connect",
		Usage: "connection to kubernetes cluster",
		Flags: ConnectActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := completeOptions(options); err != nil {
				return err
			}
			if err := combineKubeOpts(options); err != nil {
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

	options.RuntimeOptions.Component = common.ComponentConnect
	err := util.WritePidFile(common.ComponentConnect)
	if err != nil {
		return err
	}
	log.Info().Msgf("KtConnect start at %d", os.Getpid())

	ch := SetUpCloseHandler(cli, options, common.ComponentConnect)
	if err = connectToCluster(cli, options); err != nil {
		return err
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted: %s", <-process.Interrupt())
		CleanupWorkspace(cli, options)
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

func connectToCluster(cli kt.CliInterface, options *options.DaemonOptions) (err error) {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return
	}

	if util.IsWindows() || len(options.ConnectOptions.Dump2HostsNamespaces) > 0 {
		setupDump2Host(options, kubernetes)
	}
	if options.ConnectOptions.Method == common.ConnectMethodVpn {
		checkSshuttleInstalled(cli.Exec().Sshuttle())
	} else if options.ConnectOptions.Method == common.ConnectMethodSocks {
		setupGlobalProxy(options)
	}

	endPointIP, podName, credential, err := getOrCreateShadow(options, err, kubernetes)
	if err != nil {
		return
	}

	cidrs, err := kubernetes.ClusterCidrs(options.Namespace, options.ConnectOptions)
	if err != nil {
		return
	}

	return cli.Shadow().Outbound(podName, endPointIP, credential, cidrs, cli.Exec())
}

func checkSshuttleInstalled(cli sshuttle.CliInterface) {
	if !exec.CanRun(cli.Version()) {
		err := exec.RunAndWait(cli.Install(), "install_sshuttle")
		if err != nil {
			log.Error().Msgf("Failed find or install sshuttle: %s", err)
		}
	}
}

func setupGlobalProxy(options *options.DaemonOptions) {
	err := registry.SetGlobalProxy(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
	if err != nil {
		log.Error().Msgf("Unable to setup global connect proxy: %s, if you are not administrator,"+
			" please consider use '--method=socks5' parameter", err.Error())
	}
	err = registry.SetHttpProxyEnvironmentVariable(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
	if err != nil {
		log.Error().Msgf("Unable to setup global http proxy: %s", err.Error())
	}
}

func getOrCreateShadow(options *options.DaemonOptions, err error, kubernetes cluster.KubernetesInterface) (string, string, *util.SSHCredential, error) {
	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.ShareShadow {
		workload = fmt.Sprintf("kt-connect-daemon-connect-shared")
	}

	annotations := make(map[string]string)
	endPointIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(workload, options, labels(workload, options), annotations, envs(options))
	if err != nil {
		return "", "", nil, err
	}

	// record shadow name will clean up terminal
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	return endPointIP, podName, credential, nil
}

func setupDump2Host(options *options.DaemonOptions, kubernetes cluster.KubernetesInterface) {
	var namespaceToDump = options.ConnectOptions.Dump2HostsNamespaces
	if len(namespaceToDump) == 0 {
		namespaceToDump = append(namespaceToDump, options.Namespace)
	}
	hosts := map[string]string{}
	for _, namespace := range namespaceToDump {
		log.Debug().Msgf("Search service in %s namespace...", namespace)
		singleHosts := kubernetes.ServiceHosts(namespace)
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
	util.DumpHosts(hosts)
	options.RuntimeOptions.Dump2Host = true
}

func envs(options *options.DaemonOptions) map[string]string {
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

func labels(workload string, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentConnect,
		common.KTName:      workload,
	}
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}
	splits := strings.Split(workload, "-")
	labels[common.KTVersion] = splits[len(splits)-1]
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
