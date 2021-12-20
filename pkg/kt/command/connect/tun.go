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

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	podIP, podName, credential, err := getOrCreateShadow(kubernetes, options)
	if err != nil {
		return err
	}

	cidrs, err := kubernetes.ClusterCidrs(context.TODO(), options.Namespace, options.ConnectOptions)
	if err != nil {
		return err
	}

	stop, rootCtx, err := tunnel.ForwardSSHTunnelToLocal(cli.Exec().PortForward(), cli.Exec().Kubectl(),
		options, podName, options.ConnectOptions.SSHPort)
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
		options.ConnectOptions.SourceIP = srcIP
		options.ConnectOptions.DestIP = destIP
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

	// 1. Create tun device.
	err = cli.Tunnel().AddDevice()
	if err != nil {
		return err
	}
	log.Info().Msgf("Add tun device successful")

	// 2. Setup device ip
	err = cli.Tunnel().SetDeviceIP()
	if err != nil {
		return err
	}
	log.Info().Msgf("Set tun device ip successful")

	// 3. Create ssh tunnel.
	err = util.BackgroundRunWithCtx(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  cli.SSH().TunnelToRemote(0, credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort),
		Name: "ssh_tun",
		Stop: stop,
	})

	if err != nil {
		return err
	} else {
		log.Info().Msgf("Create ssh tun successful")
	}

	// 4. Add route to kubernetes cluster.
	for i := range cidrs {
		err = cli.Tunnel().AddRoute(cidrs[i])
		if err != nil {
			return err
		}
		log.Info().Msgf("Add route %s successful", cidrs[i])
	}

	if !options.ConnectOptions.DisableDNS {
		// 6. Setup dns config.
		// This will overwrite the file /etc/resolv.conf
		err = util.AddNameserver(podIP)
		if err == nil {
			log.Info().Msgf("Add nameserver %s successful", podIP)
		}
		return err
	}

	return nil
}
