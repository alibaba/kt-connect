package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"io/ioutil"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/channel"

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
		stop, rootCtx, err := forwardSshPortToLocal(cli, s.Options, name)
		if err == nil {
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

func forwardSshPortToLocal(cli exec.CliInterface, options *options.DaemonOptions, name string) (chan bool, context.Context, error) {
	stop := make(chan bool)
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func(wg *sync.WaitGroup) {
		err = exec.BackgroundRunWithCtx(&exec.CMDContext{
			Ctx: rootCtx,
			Cmd: cli.Kubectl().PortForward(
				options.Namespace,
				name,
				common.SshPort,
				options.ConnectOptions.SSHPort),
			Name: "port-forward",
			Stop: stop,
		})
		util.WaitPortBeReady(options.WaitTime, options.ConnectOptions.SSHPort)
		wg.Done()
	}(&wg)

	wg.Wait()
	if err != nil {
		return nil, nil, err
	}
	return stop, rootCtx, err
}

func startSocks4Connection(cli exec.CliInterface, options *options.DaemonOptions, name string) error {
	showSocksBanner(common.ConnectMethodSocks, options.ConnectOptions.SocksPort)
	return cli.Kubectl().PortForward(options.Namespace, name, common.Socks4Port, options.ConnectOptions.SocksPort).Start()
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

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan bool) (err error) {
	err = exec.BackgroundRunWithCtx(&exec.CMDContext{
		Ctx: rootCtx,
		Cmd: cli.SSHUttle().Connect(credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort,
			podIP, options.ConnectOptions.DisableDNS, cidrs, options.Debug),
		Name: "vpn(sshuttle)",
		Stop: stop,
	})
	return err
}
