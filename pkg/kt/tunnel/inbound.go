package tunnel

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"strings"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// ForwardPodToLocal mapping pod port to local port
func ForwardPodToLocal(exposePorts, podName, privateKey string, opt *options.DaemonOptions) (int, error) {
	log.Info().Msgf("Forwarding pod %s to local via port %s", podName, exposePorts)
	localSSHPort := util.GetRandomSSHPort()
	if localSSHPort < 0 {
		return -1, fmt.Errorf("failed to find any available local port")
	}

	// port forward pod 22 -> local <random port>
	_, _, err := ForwardSSHTunnelToLocal(opt, podName, localSSHPort)
	if err != nil {
		return -1, err
	}

	if opt.ExchangeOptions.Mode != common.ExchangeModeEphemeral {
		// remote forward pod -> local via ssh
		var wg sync.WaitGroup
		ssh := sshchannel.SSHChannel{}
		// supports multi port pairs
		portPairs := strings.Split(exposePorts, ",")
		for _, exposePort := range portPairs {
			localPort, remotePort, err2 := util.ParsePortMapping(exposePort)
			if err2 != nil {
				return -1, err2
			}
			ExposeLocalPort(&wg, &ssh, localPort, remotePort, localSSHPort, privateKey)
		}
		wg.Wait()
	}
	return localSSHPort, nil
}

// ExposeLocalPort forward remote pod to local
func ExposeLocalPort(wg *sync.WaitGroup, ssh sshchannel.Channel, localPort, remotePort, localSSHPort int, privateKey string) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Debug().Msgf("Exposing remote pod:%d to local port localhost:%d", remotePort, localPort)
		err := ssh.ForwardRemoteToLocal(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", localSSHPort),
			fmt.Sprintf("0.0.0.0:%d", remotePort),
			fmt.Sprintf("127.0.0.1:%d", localPort),
		)
		if err != nil {
			log.Error().Err(err).Msgf("Error happen when forward remote request to local")
		}
		wg.Done()
	}(wg)
}
