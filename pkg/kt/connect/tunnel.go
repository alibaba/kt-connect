package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/portforward"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/resolvconf"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

func forwardSSHTunnelToLocal(cli portforward.CliInterface, options *options.DaemonOptions, podName string, localSSHPort int) (chan struct{}, context.Context, error) {
	stop, rootCtx, err := cli.ForwardPodPortToLocal(portforward.Request{
		RestConfig: options.RuntimeOptions.RestConfig,
		PodName:    podName,
		Namespace:  options.Namespace,
		PodPort:    common.SshPort,
		LocalPort:  localSSHPort,
		Timeout:    options.WaitTime,
	})
	return stop, rootCtx, err
}

func forwardSocksTunnelToLocal(cli portforward.CliInterface, options *options.DaemonOptions, podName string) error {
	showSetupSuccessfulMessage(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	_, _, err := cli.ForwardPodPortToLocal(portforward.Request{
		RestConfig: options.RuntimeOptions.RestConfig,
		PodName:    podName,
		Namespace:  options.Namespace,
		PodPort:    common.Socks4Port,
		LocalPort:  options.ConnectOptions.SocksPort,
		Timeout:    options.WaitTime,
	})
	return err
}

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions) (err error) {
	showSetupSuccessfulMessage(common.ConnectMethodSocks5, options.ConnectOptions.SocksPort)
	_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
		options.ConnectOptions.SocksPort)), 0644)

	return ssh.StartSocks5Proxy(
		&sshchannel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
	)
}

func showSetupSuccessfulMessage(protocol string, port int) {
	log.Info().Msgf("Start %s proxy successfully", protocol)
	if !util.IsWindows() {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Please setup proxy config by: export http_proxy=%s://127.0.0.1:%d", protocol, port)
		log.Info().Msgf("==============================================================")
	}
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
	err = exec.RunAndWait(cli.Tunnel().AddDevice(), "add_device", options.Debug)
	if err != nil {
		return err
	} else {
		log.Info().Msgf("Add tun device successful")
	}

	// 2. Setup device ip
	err = exec.RunAndWait(cli.Tunnel().SetDeviceIP(), "set_device_ip", options.Debug)
	if err != nil {
		// clean up
		return err
	} else {
		log.Info().Msgf("Set tun device ip successful")
	}

	// 3. Set device up.
	err = exec.RunAndWait(cli.Tunnel().SetDeviceUp(), "set_device_up", options.Debug)
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
		err = exec.RunAndWait(cli.Tunnel().AddRoute(cidrs[i]), "add_route", options.Debug)
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
