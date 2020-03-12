package connect

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/options"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster
func (s *Shadow) Inbound(exposePort, podName, remoteIP string, credential *util.SSHCredential) (err error) {
	kubernetesCli := &kubectl.Cli{
		KubeConfig: s.Options.KubeConfig,
	}
	sshCli := &ssh.Cli{}
	return inbound(exposePort, podName, remoteIP, credential, s.Options, kubernetesCli, sshCli)
}

func inbound(
	exposePort, podName, remoteIP string, credential *util.SSHCredential,
	options *options.DaemonOptions,
	kubernetesCli kubectl.CliInterface,
	sshCli ssh.CliInterface,
) (err error) {
	debug := options.Debug
	namespace := options.Namespace

	log.Info().Msgf("remote %s forward to local %s", remoteIP, exposePort)
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		portforward := kubernetesCli.PortForward(namespace, podName, localSSHPort)
		err = exec.BackgroundRun(portforward, "exchange port forward to local", debug)
		// make sure port-forward already success
		time.Sleep(time.Duration(2) * time.Second)
		wg.Done()
	}(&wg)
	wg.Wait()
	if err != nil {
		return
	}
	log.Info().Msgf("redirect request from pod %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	localPort := exposePort
	remotePort := exposePort
	ports := strings.SplitN(exposePort, ":", 2)
	if len(ports) > 1 {
		localPort = ports[1]
		remotePort = ports[0]
	}
	cmd := sshCli.ForwardRemoteRequestToLocal(localPort, credential.RemoteHost, remotePort, credential.PrivateKeyPath, localSSHPort)
	return exec.BackgroundRun(cmd, "ssh remote port-forward", debug)
}
