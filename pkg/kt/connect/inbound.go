package connect

import (
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster
func (s *Shadow) Inbound(exposePort, podName, remoteIP string) (err error) {
	debug := s.Options.Debug
	kubeConfig := s.Options.KubeConfig
	namespace := s.Options.Namespace
	log.Info().Msgf("remote %s forward to local %s", remoteIP, exposePort)
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		portforward := kubectl.PortForward(kubeConfig, namespace, podName, localSSHPort)
		err = exec.BackgroundRun(portforward, "exchange port forward to local", debug)
		wg.Done()
	}(&wg)
	wg.Wait()
	if err != nil {
		return
	}
	log.Info().Msgf("SSH Remote port-forward POD %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	localPort := exposePort
	remotePort := exposePort
	ports := strings.SplitN(exposePort, ":", 2)
	if len(ports) > 1 {
		localPort = ports[1]
		remotePort = ports[0]
	}
	cmd := ssh.ForwardRemoteRequestToLocal(localPort, "127.0.0.1", remotePort, localSSHPort)
	return exec.BackgroundRun(cmd, "ssh remote port-forward", debug)
}
