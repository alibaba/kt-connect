package connect

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/portforward"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

func forwardSSHTunnelToLocal(cli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string, localSSHPort int) (stop chan struct{}, rootCtx context.Context, err error) {
	if options.UseKubectl {
		err = portForwardViaKubectl(kubectlCli, options, podName, common.SshPort, localSSHPort)
	} else {
		stop, rootCtx, err = cli.ForwardPodPortToLocal(options, podName, common.SshPort, localSSHPort)
	}
	return stop, rootCtx, err
}

func forwardSocksTunnelToLocal(pfCli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string) (err error) {
	showSetupSocksMessage(common.ConnectMethodSocks, options.ConnectOptions)
	if options.UseKubectl {
		err = portForwardViaKubectl(kubectlCli, options, podName, common.Socks4Port, options.ConnectOptions.SocksPort)
	} else {
		_, _, err = pfCli.ForwardPodPortToLocal(options, podName, common.Socks4Port, options.ConnectOptions.SocksPort)
	}
	return err
}

func portForwardViaKubectl(kubectlCli kubectl.CliInterface, options *options.DaemonOptions, podName string, remotePort, localPort int) error {
	command := kubectlCli.PortForward(options.Namespace, podName, remotePort, localPort)

	// If localSSHPort is in use by another process, return an error.
	ready := util.WaitPortBeReady(1, localPort)
	if ready {
		return fmt.Errorf("127.0.0.1:%d already in use", localPort)
	}

	err := util.BackgroundRun(command, fmt.Sprintf("forward %d to localhost:%d", remotePort, localPort))
	if err == nil {
		if !util.WaitPortBeReady(options.WaitTime, localPort) {
			err = errors.New("connect to port-forward failed")
		}
		util.SetupPortForwardHeartBeat(localPort)
	}
	return err
}

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions) (err error) {
	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		ioutil.WriteFile(jvmrcFilePath, []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
			options.ConnectOptions.SocksPort)), 0644)
	}

	showSetupSocksMessage(common.ConnectMethodSocks5, options.ConnectOptions)
	return ssh.StartSocks5Proxy(
		&sshchannel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("%s:%d", options.ConnectOptions.SocksAddr, options.ConnectOptions.SocksPort),
	)
}

func showSetupSocksMessage(protocol string, connectOptions *options.ConnectOptions) {
	port := connectOptions.SocksPort
	log.Info().Msgf("Starting up %s proxy ...", protocol)
	if !connectOptions.UseGlobalProxy {
		log.Info().Msgf("--------------------------------------------------------------")
		if util.IsWindows() {
			if util.IsCmd() {
				log.Info().Msgf("Please setup proxy config by: set http_proxy=%s://127.0.0.1:%d", protocol, port)
			} else {
				log.Info().Msgf("Please setup proxy config by: $env:http_proxy=\"%s://127.0.0.1:%d\"", protocol, port)
			}
		} else {
			log.Info().Msgf("Please setup proxy config by: export http_proxy=%s://127.0.0.1:%d", protocol, port)
		}
		log.Info().Msgf("--------------------------------------------------------------")
	}
}

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, opt *options.ConnectOptions, req *sshuttle.SSHVPNRequest) (err error) {
	err = util.BackgroundRunWithCtx(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  cli.Sshuttle().Connect(opt, req),
		Name: "vpn(sshuttle)",
		Stop: req.Stop,
	})
	return err
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
