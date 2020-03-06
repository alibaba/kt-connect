package connect

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(name, podIP string, cidrs []string) (err error) {
	options := s.Options
	err = util.PrepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = exec.BackgroundRun(
		kubectl.PortForward(
			options.KubeConfig,
			options.Namespace,
			name,
			options.ConnectOptions.SSHPort),
		"port-forward",
		options.Debug)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	if options.ConnectOptions.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", options.ConnectOptions.Socke5Proxy)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", options.ConnectOptions.Socke5Proxy)), 0644)
		err = exec.BackgroundRun(ssh.DynamicForwardLocalRequestToRemote("127.0.0.1", options.ConnectOptions.SSHPort, options.ConnectOptions.Socke5Proxy), "vpn(ssh)", options.Debug)
	} else {
		err = exec.BackgroundRun(sshuttle.SSHUttle("127.0.0.1", options.ConnectOptions.SSHPort, podIP, options.ConnectOptions.DisableDNS, cidrs, options.Debug), "vpn(sshuttle)", options.Debug)
	}
	if err != nil {
		return
	}

	log.Info().Msgf("KT proxy start successful")
	return
}
