package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/cilium/ipam/service/allocator"
	"github.com/cilium/ipam/service/ipallocator"
	"github.com/rs/zerolog/log"
	"net"
)

func ByTun(cli kt.CliInterface, options *options.DaemonOptions) error {
	if err := completeOptions(options); err != nil {
		return err
	}

	podIP, podName, credential, err := getOrCreateShadow(cli.Kubernetes(), options)
	if err != nil {
		return err
	}

	cidrs, err := cli.Kubernetes().ClusterCidrs(context.TODO(), options.Namespace, options.ConnectOptions)
	if err != nil {
		return err
	}

	stop, rootCtx, err := tunnel.ForwardSSHTunnelToLocal(cli.Exec().Kubectl(), options, podName, options.ConnectOptions.SSHPort)
	if err != nil {
		return err
	}
	return startTunConnection(rootCtx, cli.Exec(), credential, options, podIP, cidrs, stop)
}

func completeOptions(options *options.DaemonOptions) error {
	if options.ConnectOptions.Method == common.ConnectMethodTun {
		srcIP, destIP, err := allocateTunIP(options.ConnectOptions.TunCidr)
		if err != nil {
			return err
		}
		options.RuntimeOptions.SourceIP = srcIP
		options.RuntimeOptions.DestIP = destIP
	}
	return nil
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

// startTunConnection creates a ssh tunnel to pod
func startTunConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan struct{}) (err error) {

	// Create tun device.
	if err = cli.Tunnel().AddDevice(); err != nil {
		return
	}
	log.Info().Msgf("Add tun device successful")

	// Setup device ip
	if err = cli.Tunnel().SetDeviceIP(); err != nil {
		return
	}
	log.Info().Msgf("Set tun device ip successful")

	// Create ssh tunnel.
	if err = util.BackgroundRunWithCtx(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  cli.SSH().TunnelToRemote(0, credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort),
		Name: "ssh_tun",
		Stop: stop,
	}); err != nil {
		return
	} else {
		log.Info().Msgf("Create ssh tun successful")
	}

	// Add route to kubernetes cluster.
	for i := range cidrs {
		if err = cli.Tunnel().AddRoute(cidrs[i]); err != nil {
			return
		}
		log.Info().Msgf("Add route %s successful", cidrs[i])
	}

	if !options.ConnectOptions.DisableDNS {
		// Setup dns config.
		// This will overwrite the file /etc/resolv.conf
		if err = util.AddNameserver(podIP); err == nil {
			log.Info().Msgf("Add nameserver %s successful", podIP)
		}
	}

	return
}
