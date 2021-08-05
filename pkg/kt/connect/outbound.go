package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/resolvconf"
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
	switch s.Options.ConnectOptions.Method {
	case common.ConnectMethodSocks:
		err = startSocks4Connection(cli, s.Options, name)
	case common.ConnectMethodTun:
		stop, rootCtx, err := forwardSshPortToLocal(cli, s.Options, name)
		if err == nil {
			err = startTunConnection(rootCtx, cli, credential, s.Options, podIP, cidrs, stop)
		}
	default:
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

// startTunConnection creates a ssh tunnel to pod
func startTunConnection(rootCtx context.Context, cli exec.CliInterface, credential *util.SSHCredential,
	options *options.DaemonOptions, podIP string, cidrs []string, stop chan bool) (err error) {

	// 1. Create tun device.
	err = exec.RunAndWait(cli.SSHTunnelling().AddDevice(), "add_device", options.Debug)
	if err != nil {
		return err
	} else {
		log.Info().Msgf("Add tun device successful")
	}

	// 2. Setup device ip
	err = exec.RunAndWait(cli.SSHTunnelling().SetDeviceIP(), "set_device_ip", options.Debug)
	if err != nil {
		// clean up
		return err
	} else {
		log.Info().Msgf("Set tun device ip successful")
	}

	// 3. Set device up.
	err = exec.RunAndWait(cli.SSHTunnelling().SetDeviceUp(), "set_device_up", options.Debug)
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
		err = exec.RunAndWait(cli.SSHTunnelling().AddRoute(cidrs[i]), "add_route", options.Debug)
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
