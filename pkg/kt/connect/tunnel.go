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
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/resolvconf"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

func forwardSSHTunnelToLocal(cli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string, localSSHPort int) (stop chan struct{}, rootCtx context.Context, err error) {
	if options.UseKubectl {
		command := kubectlCli.PortForward(options.Namespace, podName, common.SshPort, localSSHPort)
		err = exec.BackgroundRun(command, "forward ssh to localhost")
		if err == nil {
			if !util.WaitPortBeReady(options.WaitTime, localSSHPort) {
				err = errors.New("connect to port-forward failed")
			}
		}
	} else {
		stop, rootCtx, err = cli.ForwardPodPortToLocal(portforward.Request{
			RestConfig: options.RuntimeOptions.RestConfig,
			PodName:    podName,
			Namespace:  options.Namespace,
			PodPort:    common.SshPort,
			LocalPort:  localSSHPort,
			Timeout:    options.WaitTime,
		})
	}
	return stop, rootCtx, err
}

func forwardSocksTunnelToLocal(pfCli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string) (err error) {
	showSetupSocksMessage(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	if options.UseKubectl {
		command := kubectlCli.PortForward(options.Namespace, podName, common.Socks4Port, options.ConnectOptions.SocksPort)
		err = exec.BackgroundRun(command, "forward socks to localhost")
		if err == nil {
			if !util.WaitPortBeReady(options.WaitTime, options.ConnectOptions.SocksPort) {
				err = errors.New("connect to port-forward failed")
			}
		}
	} else {
		_, _, err = pfCli.ForwardPodPortToLocal(portforward.Request{
			RestConfig: options.RuntimeOptions.RestConfig,
			PodName:    podName,
			Namespace:  options.Namespace,
			PodPort:    common.Socks4Port,
			LocalPort:  options.ConnectOptions.SocksPort,
			Timeout:    options.WaitTime,
		})
	}
	return err
}

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions) (err error) {
	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		ioutil.WriteFile(jvmrcFilePath, []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
			options.ConnectOptions.SocksPort)), 0644)
	}

	showSetupSocksMessage(common.ConnectMethodSocks5, options.ConnectOptions.SocksPort)
	return ssh.StartSocks5Proxy(
		&sshchannel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
	)
}

func showSetupSocksMessage(protocol string, port int) {
	log.Info().Msgf("Starting up %s proxy ...", protocol)
	if util.IsWindows() && protocol == common.ConnectMethodSocks {
		// socks method in windows will auto setup global proxy config
		return
	}
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

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, request SSHVPNRequest) (err error) {
	err = exec.BackgroundRunWithCtx(&exec.CMDContext{
		Ctx: rootCtx,
		Cmd: cli.Sshuttle().Connect(request.RemoteSSHHost, request.RemoteSSHPKPath, request.RemoteSSHPort,
			request.RemoteDNSServerAddress, request.DisableDNS, request.CustomCRID, request.Debug),
		Name: "vpn(sshuttle)",
		Stop: request.Stop,
	})
	return err
}

// startTunConnection creates a ssh tunnel to pod
func startTunConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan struct{}) (err error) {

	// 1. Create tun device.
	err = exec.RunAndWait(cli.Tunnel().AddDevice(), "add_device")
	if err != nil {
		return err
	} else {
		log.Info().Msgf("Add tun device successful")
	}

	// 2. Setup device ip
	err = exec.RunAndWait(cli.Tunnel().SetDeviceIP(), "set_device_ip")
	if err != nil {
		// clean up
		return err
	} else {
		log.Info().Msgf("Set tun device ip successful")
	}

	// 3. Set device up.
	err = exec.RunAndWait(cli.Tunnel().SetDeviceUp(), "set_device_up")
	if err != nil {
		return err
	} else {
		log.Info().Msgf("Set tun device up successful")
	}

	// 4. Create ssh tunnel.
	err = exec.BackgroundRunWithCtx(&exec.CMDContext{
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

	// 5. Add route to kubernetes cluster.
	for i := range cidrs {
		err = exec.RunAndWait(cli.Tunnel().AddRoute(cidrs[i]), "add_route")
		if err != nil {
			// clean up
			return err
		} else {
			log.Info().Msgf("Add route %s successful", cidrs[i])
		}
	}

	if !options.ConnectOptions.DisableDNS {
		// 6. Setup dns config.
		// This will overwrite the file /etc/resolv.conf
		err = (&resolvconf.Conf{}).AddNameserver(podIP)
		if err == nil {
			log.Info().Msgf("Add nameserver %s successful", podIP)
		}
		return err
	}

	return nil
}
