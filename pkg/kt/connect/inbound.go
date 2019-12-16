package connect

import (
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// remotePortForward mapping local port from cluster
func remotePortForward(expose string, kubeconfig string, namespace string, target string, remoteIP string, debug bool) (err error) {
	log.Info().Msgf("remote %s forward to local %s", remoteIP, expose)
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
