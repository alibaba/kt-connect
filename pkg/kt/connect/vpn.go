package connect

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"k8s.io/client-go/kubernetes"
)

// GetProxyCrids get All VPN CRID
func (c *Connect) GetProxyCrids(clientset *kubernetes.Clientset) (cidrs []string, err error) {
	cidrs, err = util.GetCirds(clientset, c.PodCIDR)
	return
}

// StartConnect start vpn connection
func (c *Connect) StartConnect(name string, podIP string, cidrs []string) (err error) {
	err = util.PrepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = util.BackgroundRun(util.PortForward(c.Options.KubeConfig, c.Options.Namespace, name, c.Options.ConnectOptions.SSHPort), "port-forward", c.Options.Debug)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	if c.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", c.Options.ConnectOptions.Socke5Proxy)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", c.Options.ConnectOptions.Socke5Proxy)), 0644)
		err = util.BackgroundRun(util.SSHDynamicPortForward("127.0.0.1", c.Options.ConnectOptions.SSHPort, c.Options.ConnectOptions.Socke5Proxy), "vpn(ssh)", c.Options.Debug)
	} else {
		err = util.BackgroundRun(util.SSHUttle("127.0.0.1", c.Options.ConnectOptions.SSHPort, podIP, c.Options.ConnectOptions.DisableDNS, cidrs, c.Options.Debug), "vpn(sshuttle)", c.Debug)
	}
	if err != nil {
		return
	}

	log.Printf("KT proxy start successful")
	return
}
