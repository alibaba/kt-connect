package connect

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/alibaba/kt-connect/internal/network"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/channel"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

func forwardSSHTunnelToLocal(s *Shadow, podName string, localSSHPort int) (chan struct{}, context.Context, error) {
	stop, rootCtx, err := network.ForwardPodPortToLocal(network.PortForwardAPodRequest{
		RestConfig: s.Options.RuntimeOptions.RestConfig,
		PodName:    podName,
		Namespace:  s.Options.Namespace,
		PodPort:    common.SshPort,
		LocalPort:  localSSHPort,
		Timeout:    s.Options.WaitTime,
	})
	return stop, rootCtx, err
}

func forwardSocksTunnelToLocal(options *options.DaemonOptions, podName string) error {
	showSocksBanner(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	_, _, err := network.ForwardPodPortToLocal(network.PortForwardAPodRequest{
		RestConfig: options.RuntimeOptions.RestConfig,
		PodName:    podName,
		Namespace:  options.Namespace,
		PodPort:    common.Socks4Port,
		LocalPort:  options.ConnectOptions.SocksPort,
		Timeout:    options.WaitTime,
	})
	return err
}

func startSocks5Connection(ssh channel.Channel, options *options.DaemonOptions) (err error) {
	showSocksBanner(common.ConnectMethodSocks5, options.ConnectOptions.SocksPort)
	_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
		options.ConnectOptions.SocksPort)), 0644)
	_ = ioutil.WriteFile(".envrc", []byte(fmt.Sprintf("KUBERNETES_NAMESPACE=%s",
		options.Namespace)), 0644)

	return ssh.StartSocks5Proxy(
		&channel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
	)
}

func showSocksBanner(protocol string, port int) {
	operation := "export"
	if util.IsWindows() {
		operation = "set"
	}
	log.Info().Msgf("==============================================================")
	log.Info().Msgf("Start SOCKS Proxy Successful: %s http_proxy=%s://127.0.0.1:%d", operation, protocol, port)
	log.Info().Msgf("==============================================================")
}

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, request SSHVPNRequest) (err error) {
	err = exec.BackgroundRunWithCtx(&exec.CMDContext{
		Ctx: rootCtx,
		Cmd: cli.SSHUttle().Connect(request.RemoteSSHHost, request.RemoteSSHPKPath, request.RemoteSSHPort,
			request.RemoteDNSServerAddress, request.DisableDNS, request.CustomCRID, request.Debug),
		Name: "vpn(sshuttle)",
		Stop: request.Stop,
	})
	return err
}
