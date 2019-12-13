package connect

import (
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

var (
	pk []byte
)

func init() {
	rand.Seed(time.Now().UnixNano())
	pk = []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEpAIBAAKCAQEAvSVAezJDBhrNDhLhuCaKrvdtCFTdqJmGLGyfBqEYb3p4a91g\n" +
		"l4gD2LGwiRlgpwU4oSECeMmwP53C4vrfPKY45/+8lncIkE/4E+8hHssXkHaqQrjE\n" +
		"FtePfxZ6/xi84kUbWNNV4IGAeFwXtq9GszQ+kWMNT5QmuexOXOlqq7W4CIAUe3uX\n" +
		"29WCp3OGiBeP4ORDraRa/1bwBH+Cq0UxEYT+6EuDU0YzF3JF4H8At6NdgElAuezE\n" +
		"wI84p5LNr1HmTPndcHJtX2+POKEoNYBxPekEyJbqExIR2dLRUytlX5tIacKkMBCJ\n" +
		"aX7DBtJzWX7BxfjRfXjzNOpufuTU9BsknyJlFwIDAQABAoIBAQCS606s4xvAsCy7\n" +
		"U9tUyUtMIRDmOdV7UtUvyKe15Igwf3bugiS3T4V9Wnh/5eB3m8yjDBr5a+ClaYup\n" +
		"96hTWeI2AyWf0pIqVpOiGEsnuiVxp1sVPKPEAmiKFRIw+CwvrfJSCsZX/v+lfhNF\n" +
		"adyG8nvvPntmZvO102IDNaQQALUUk+J69yvtb7ekfvZCSmanXx/R7y2+u8Hd8wtf\n" +
		"fpyVcM1g7YZwhxto2uKyUE1T/myOT5+wULfYTMNynsLB6dXJJP+a89I6cRdFCIib\n" +
		"kqoZ5FuaTXXucrgWGDne6DcvNbiZi1f8LRb6RFnlvv6D42xyjelyoGY7BktKsFwJ\n" +
		"NwLR1lBhAoGBANyGB7ti190DZoDGQvIpSHd+JeZHoK3g49VNxGFU+SAhXa8gQn6K\n" +
		"Xi5qNRD2XLTEnT36U20/bkcDv0oSTZikJhU0OqxgouVO3YZ2cZhXcoPZmf+vhgzI\n" +
		"ufv0T8/HlQyr7Sp9VqfqlC7u1P83VbNro/D3V+wtNKog4g++DvKtsh8JAoGBANuS\n" +
		"9XDaAnq5K0rjShNqCMRyHx6+kFPaLpL4kt1f4yra8w/m3pes63sO1vz/4wOhoalL\n" +
		"imAEqTKTblinPhjCxbe4e/WqnAQM05XROdiGer2RhBIMCo2/YE3WLCWAyVCDtd3B\n" +
		"Te9rPynSsAmtgDRuftusY7TAIuwZuG4K71Gw8UsfAoGAYmRm5MPYXqNaw9AyJIwo\n" +
		"6i/dxx5kYdB6tzxoh6j7MsvQWggBwyYHmZwHq1bQzFMBeZrMSG1JzeOtIOaDurxa\n" +
		"xZE1MJ45cCi9DHaifn9d99hKLtvo6qFQ4ksCpUl+hlXbjt63oFo43avwWyMcWN6J\n" +
		"GkWx9A3DdrkPREjfsIWxeMkCgYEAtDhv6duWk2IujX32y+6JGaxNrK9eyORYu9r4\n" +
		"uGi+jOs++ztUUgvlD5EDlo70poNgrBLLlbndohxuQqeqiSo8nGn4nJAXFB/u/pXH\n" +
		"M9hVIAky7JkjhGqiweBbRcDp+4LPoB7MOAm/wzUhth/JDb/vsaBSCgZ143HM9c1V\n" +
		"1qgztKMCgYAUhQRJB6ofGqiGsPN2KZw+0IoPNS3Tk0NTjzVh2o927B8zb0T0bO5e\n" +
		"qe0OO7FFGcON6uSOkGu2p9KHUEm6OFaQLjdysjrGI7GVRYW7D/SSLidRREv2A70R\n" +
		"f0/Mi8v9nD4ztroXQDeeL8O4rFTnfRdqs+MZ/MYoq9C5iE1IHJm7KQ==\n" +
		"-----END RSA PRIVATE KEY-----")
}

// Connect VPN connect interface
type Connect struct {
	Kubeconfig string
	Namespace  string
	Image      string
	Method     string
	Swap       string
	Expose     string
	ProxyPort  int
	Port       int // Local SSH Port
	DisableDNS bool
	PodCIDR    string
	Debug      bool
	PidFile    string
}

// PrepareSSHPrivateKey generator ssh private key
func (c *Connect) PrepareSSHPrivateKey() (err error) {
	privateKey := util.PrivateKeyPath()
	err = ioutil.WriteFile(privateKey, pk, 400)
	if err != nil {
		log.Printf("Fails create temp ssh private key")
	}
	return
}

func remotePortForward(expose string, kubeconfig string, namespace string, target string, remoteIP string, debug bool) (err error) {
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	portforward := util.PortForward(kubeconfig, namespace, target, localSSHPort)
	err = util.BackgroundRun(portforward, "exchange port forward to local", debug)
	if err != nil {
		return
	}

	time.Sleep(time.Duration(2) * time.Second)
	log.Printf("SSH Remote port-forward POD %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	localPort := expose
	remotePort := expose
	ports := strings.SplitN(expose, ":", 2)
	if len(ports) > 1 {
		localPort = ports[1]
		remotePort = ports[0]
	}
	cmd := util.SSHRemotePortForward(localPort, "127.0.0.1", remotePort, localSSHPort)
	return util.BackgroundRun(cmd, "ssh remote port-forward", debug)
}

