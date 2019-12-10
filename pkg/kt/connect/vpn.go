package connect

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
		_ = ioutil.WriteFile(".jvmrc", []byte(fmt.Sprintf("-Dhttp.proxyHost=127.0.0.1\n-Dhttp.proxyPort=%d", c.ProxyPort)), 0644)
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

// OnConnectExit handle connect exit
func (c *Connect) OnConnectExit(name string, pid int) {
	os.Remove(c.PidFile)
	os.Remove(".jvmrc")
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("Cleanup proxy deplyment %s", name)
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}
