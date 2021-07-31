package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/channel"
	"io/ioutil"

	"github.com/alibaba/kt-connect/pkg/kt/options"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	ssh := channel.SSHChannel{}
	return outbound(s, name, podIP, credential, cidrs, cli, &ssh)
}

func outbound(s *Shadow, name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface, ssh channel.Channel) (err error) {
	if s.Options.ConnectOptions.Method == common.ConnectMethodSocks {
		err = startSocks4Connection(cli, s.Options, name)
	} else {
		stop, rootCtx, err2 := forwardSshPortToLocal(cli, s.Options, name)
		if err2 == nil {
			if s.Options.ConnectOptions.Method == common.ConnectMethodSocks5 {
				err = startSocks5Connection(ssh, s.Options)
			} else {
				err = startVPNConnection(rootCtx, cli, credential, s.Options, podIP, cidrs, stop)
			}
		}
	}
	if err != nil {
		return
	}

	log.Info().Msgf("Proxy start successful")
	return
}

func forwardSshPortToLocal(cli exec.CliInterface, options *options.DaemonOptions, name string) (chan string, context.Context, error) {
	stop := make(chan string)
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	err := exec.BackgroundRunWithCtx(&exec.CMDContext{
		Ctx: rootCtx,
		Cmd: cli.Kubectl().PortForward(
			options.Namespace,
			name,
			common.SshPort,
			options.ConnectOptions.SSHPort),
		Name: "port-forward",
		Stop: stop,
	})
	if err != nil {
		return nil, nil, err
	}
	util.WaitPortBeReady(options.WaitTime, options.ConnectOptions.SSHPort)
	return stop, rootCtx, nil
}

func startSocks4Connection(cli exec.CliInterface, options *options.DaemonOptions, name string) (err error) {
	err = cli.Kubectl().PortForward(options.Namespace, name, common.Socks4Port, options.ConnectOptions.SocksPort).Start()
	if err == nil {
		showSetupSuccessfulMessage(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	}
	return err
}

func startSocks5Connection(ssh channel.Channel, options *options.DaemonOptions) (err error) {
	_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
		options.ConnectOptions.SocksPort)), 0644)

	err = ssh.StartSocks5Proxy(
		&channel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
	)
	if err == nil {
		showSetupSuccessfulMessage(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	}
	return err
}

func showSetupSuccessfulMessage(protocol string, port int) {
	log.Info().Msgf("Start %s proxy successfully", protocol)
	if !util.IsWindows() {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Please setup proxy config by: export http_proxy=%s://127.0.0.1:%d", protocol, port)
		log.Info().Msgf("==============================================================")
	}
}

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan string) (err error) {
	return exec.BackgroundRunWithCtx(&exec.CMDContext{
		Ctx: rootCtx,
		Cmd: cli.SSHUttle().Connect(credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort,
			podIP, options.ConnectOptions.DisableDNS, cidrs, options.Debug),
		Name: "vpn(sshuttle)",
		Stop: stop,
	})
}
