package connect

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster
func (s *Shadow) Inbound(exposePorts, podName, remoteIP string, _ *util.SSHCredential) (err error) {
	log.Info().Msg("Creating shadow inbound(remote->local)")
	return inbound(s, exposePorts, podName, remoteIP, &sshchannel.SSHChannel{}, &exec.Cli{})
}

func inbound(s *Shadow, exposePorts, podName, remoteIP string, ssh sshchannel.Channel, cli exec.CliInterface) (err error) {
	log.Info().Msgf("Remote %s forward to local %s", remoteIP, exposePorts)
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}

	_, _, err = forwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, localSSHPort)
	if err != nil {
		return
	}

	exposeLocalPorts(ssh, exposePorts, localSSHPort)
	return nil
}

func exposeLocalPorts(ssh sshchannel.Channel, exposePorts string, localSSHPort int) {
	var wg sync.WaitGroup
	// supports multi port pairs
	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		exposeLocalPort(&wg, ssh, exposePort, localSSHPort)
	}
	wg.Wait()
}

func exposeLocalPort(wg *sync.WaitGroup, ssh sshchannel.Channel, exposePort string, localSSHPort int) {
	localPort, remotePort := getPortMapping(exposePort)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Info().Msgf("ExposeLocalPortsToRemote request from pod:%s to 127.0.0.1:%s", remotePort, localPort)
		err := ssh.ForwardRemoteToLocal(
			&sshchannel.Certificate{
				Username: "root",
				Password: "root",
			},
			fmt.Sprintf("127.0.0.1:%d", localSSHPort),
			fmt.Sprintf("0.0.0.0:%s", remotePort),
			fmt.Sprintf("127.0.0.1:%s", localPort),
		)
		if err != nil {
			log.Error().Msgf("Error happen when forward remote request to local %s", err)
		}
		log.Info().Msgf("ExposeLocalPortsToRemote request from pod:%s to 127.0.0.1:%s finished", remotePort, localPort)
		wg.Done()
	}(wg)
}

func getPortMapping(exposePort string) (string, string) {
	localPort := exposePort
	remotePort := exposePort
	ports := strings.SplitN(exposePort, ":", 2)
	if len(ports) > 1 {
		localPort = ports[0]
		remotePort = ports[1]
	}
	return localPort, remotePort
}
