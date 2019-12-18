package connect

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// StartConnect start vpn connection
func StartConnect(name string, podIP string, cidrs []string, options *options.DaemonOptions) (err error) {
	debug := options.Debug
	err = util.PrepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = util.BackgroundRun(util.PortForward(options.KubeConfig, options.Namespace, name, options.ConnectOptions.SSHPort), "port-forward", debug)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	if options.ConnectOptions.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", options.ConnectOptions.Socke5Proxy)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", options.ConnectOptions.Socke5Proxy)), 0644)
		err = util.BackgroundRun(util.SSHDynamicPortForward("127.0.0.1", options.ConnectOptions.SSHPort, options.ConnectOptions.Socke5Proxy), "vpn(ssh)", debug)
	} else {
		err = util.BackgroundRun(util.SSHUttle("127.0.0.1", options.ConnectOptions.SSHPort, podIP, options.ConnectOptions.DisableDNS, cidrs, debug), "vpn(sshuttle)", debug)
	}
	if err != nil {
		return
	}

	log.Printf("KT proxy start successful")
	return
}
