package connect

import (
	"context"
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
	kubernetesCli := &kubectl.Cli{KubeOptions: s.Options.KubeOptions}
	sshCli := &ssh.Cli{}
	log.Info().Msg("creating shadow inbound(remote->local)")
	return inbound(exposePort, podName, remoteIP, credential, s.Options, kubernetesCli, sshCli)
}

func inbound(
	exposePorts, podName, remoteIP string, credential *util.SSHCredential,
	options *options.DaemonOptions,
	kubernetesCli kubectl.CliInterface,
	sshCli ssh.CliInterface,
) (err error) {
	debug := options.Debug
	namespace := options.Namespace
	stop := make(chan bool)
	rootCtx, cancel := context.WithCancel(context.Background())

	// one of the background process start failed and will cancel the started process
	go func() {
		util.StopBackendProcess(<-stop, cancel)
	}()

	log.Info().Msgf("remote %s forward to local %s", remoteIP, exposePorts)
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		portforward := kubernetesCli.PortForward(namespace, podName, localSSHPort)
		err = exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx:  rootCtx,
				Cmd:  portforward,
				Name: "exchange port forward to local",
				Stop: stop,
			},
			debug,
		)
		// make sure port-forward already success
		time.Sleep(time.Duration(2) * time.Second)
		wg.Done()
	}(&wg)
	wg.Wait()
	if err != nil {
		return
	}
	log.Info().Msgf("redirect request from pod %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	
	// supports multi port pairs
	portPairs := strings.SplitN(exposePorts, ",", 2)
	for _, exposePort := range portPairs {
		localPort := exposePort
		remotePort := exposePort
		ports := strings.SplitN(exposePort, ":", 2)
		if len(ports) > 1 {
			localPort = ports[1]
			remotePort = ports[0]
		}
		cmd := sshCli.ForwardRemoteRequestToLocal(localPort, credential.RemoteHost, remotePort, credential.PrivateKeyPath, localSSHPort)
		err := exec.BackgroundRunWithCtx(
			&exec.CMDContext{
				Ctx:  rootCtx,
				Cmd:  cmd,
				Name: "ssh remote port-forward",
				Stop: stop,
			},
			debug,
		)

		if err != nil {
			return err
		}
	}
	return nil
}
