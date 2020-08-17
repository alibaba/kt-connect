package connect

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	options := s.Options
	stop := make(chan bool)
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err = exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx: rootCtx,
				Cmd: cli.Kubectl().PortForward(
					options.Namespace,
					name,
					options.ConnectOptions.SSHPort),
				Name: "port-forward",
				Stop: stop,
			},
			options.Debug,
		)
		log.Info().Msgf("wait(%ds) port-forward successful", options.WaitTime)
		time.Sleep(time.Duration(options.WaitTime) * time.Second)
		wg.Done()
	}(&wg)

	wg.Wait()
	if err != nil {
		return
	}

	if options.ConnectOptions.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", options.ConnectOptions.Socke5Proxy)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", options.ConnectOptions.Socke5Proxy)), 0644)
		_ = ioutil.WriteFile(".envrc", []byte(fmt.Sprintf("KUBERNETES_NAMESPACE=%s", options.Namespace)), 0644)
		err = exec.BackgroundRunWithCtx(&exec.CMDContext{
			Ctx:  rootCtx,
			Cmd:  cli.SSH().DynamicForwardLocalRequestToRemote(credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort, options.ConnectOptions.Socke5Proxy),
			Name: "vpn(ssh)",
			Stop: stop,
		}, options.Debug)
	} else {
		err = exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx:  rootCtx,
				Cmd:  cli.SSHUttle().Connect(credential.RemoteHost, credential.PrivateKeyPath, options.ConnectOptions.SSHPort, podIP, options.ConnectOptions.DisableDNS, cidrs, options.Debug),
				Name: "vpn(sshuttle)",
				Stop: stop,
			},
			options.Debug,
		)
	}
	if err != nil {
		return
	}

	log.Info().Msgf("KT proxy start successful")
	return
}
