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
	err = c.PrepareSSHPrivateKey()
	if err != nil {
		return
	}
	err = util.BackgroundRun(util.PortForward(c.Kubeconfig, c.Namespace, name, c.Port), "port-forward", c.Debug)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(5) * time.Second)
	if c.Method == "socks5" {
		log.Info().Msgf("==============================================================")
		log.Info().Msgf("Start SOCKS5 Proxy: export http_proxy=socks5://127.0.0.1:%d", c.ProxyPort)
		log.Info().Msgf("==============================================================")
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d", c.ProxyPort)), 0644)
		err = util.BackgroundRun(util.SSHDynamicPortForward("127.0.0.1", c.Port, c.ProxyPort), "vpn(ssh)", c.Debug)
	} else {
		err = util.BackgroundRun(util.SSHUttle("127.0.0.1", c.Port, podIP, c.DisableDNS, cidrs, c.Debug), "vpn(sshuttle)", c.Debug)
	}
	if err != nil {
		return
	}

	log.Printf("KT proxy start successful")
	return
}
