package tunnel

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"strings"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster
func (s *Shadow) Inbound(exposePorts, podName string) (int, error) {
	log.Info().Msgf("Forwarding pod %s to local via port %s", podName, exposePorts)
	localSSHPort := util.GetRandomSSHPort()
	if localSSHPort < 0 {
		return -1, fmt.Errorf("failed to find any available local port")
	}

	// port forward pod 22 -> local <random port>
	cli := exec.Cli{}
	_, _, err := ForwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, localSSHPort)
	if err != nil {
		return -1, err
	}

	if s.Options.ExchangeOptions.Method != common.ExchangeMethodEphemeral {
		// remote forward pod -> local via ssh
		var wg sync.WaitGroup
		ssh := sshchannel.SSHChannel{}
		// supports multi port pairs
		portPairs := strings.Split(exposePorts, ",")
		for _, exposePort := range portPairs {
			localPort, remotePort := util.ParsePortMapping(exposePort)
			s.ExposeLocalPort(&wg, &ssh, localPort, remotePort, localSSHPort)
		}
		wg.Wait()
	}
	return localSSHPort, nil
}

func (s *Shadow) ExposeLocalPort(wg *sync.WaitGroup, ssh sshchannel.Channel, localPort, remotePort string, localSSHPort int) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Debug().Msgf("Exposing remote pod:%s to local port localhost:%s", remotePort, localPort)
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
			log.Error().Err(err).Msgf("Error happen when forward remote request to local")
		}
		wg.Done()
	}(wg)
}
