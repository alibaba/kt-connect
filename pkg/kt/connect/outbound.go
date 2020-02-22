package connect

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// StartConnect start vpn connection
func (c *Connect) StartConnect(name string, podIP string, cidrs []string, debug bool) (err error) {
	err = util.PrepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = exec.BackgroundRun(kubectl.PortForward(c.Options.KubeConfig, c.Options.Namespace, name, c.Options.ConnectOptions.SSHPort), "port-forward", debug)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	if c.Options.ConnectOptions.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", c.Options.ConnectOptions.Socke5Proxy)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", c.Options.ConnectOptions.Socke5Proxy)), 0644)
		err = exec.BackgroundRun(ssh.SSHDynamicPortForward("127.0.0.1", c.Options.ConnectOptions.SSHPort, c.Options.ConnectOptions.Socke5Proxy), "vpn(ssh)", debug)
	} else {
		err = exec.BackgroundRun(sshuttle.SSHUttle("127.0.0.1", c.Options.ConnectOptions.SSHPort, podIP, c.Options.ConnectOptions.DisableDNS, cidrs, debug), "vpn(sshuttle)", debug)
	}
	if err != nil {
		return
	}

	log.Info().Msgf("KT proxy start successful")
	return
}
