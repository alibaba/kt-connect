package transmission

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
)

// ForwardPodToLocal mapping pod port to local port
func ForwardPodToLocal(exposePorts, podName, privateKey string) (int, error) {
	log.Info().Msgf("Forwarding pod %s to local via port %s", podName, exposePorts)
	localSSHPort, err := util.GetRandomTcpPort()
	if err != nil {
		return -1, err
	}

	// port forward pod 22 -> local <random port>
	if _, err = SetupPortForwardToLocal(podName, common.StandardSshPort, localSSHPort); err != nil {
		return -1, err
	}

	if opt.Get().ExchangeOptions.Mode != util.ExchangeModeEphemeral {
		// remote forward pod -> local via ssh
		var wg sync.WaitGroup
		// supports multi port pairs
		portPairs := strings.Split(exposePorts, ",")
		for _, exposePort := range portPairs {
			localPort, remotePort, err2 := util.ParsePortMapping(exposePort)
			if err2 != nil {
				return -1, err2
			}
			ExposeLocalPort(&wg, localPort, remotePort, localSSHPort, privateKey)
		}
		wg.Wait()
	}
	return localSSHPort, nil
}

// ExposeLocalPort forward remote pod to local
func ExposeLocalPort(wg *sync.WaitGroup, localPort, remotePort, localSSHPort int, privateKey string) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Debug().Msgf("Exposing remote pod:%d to local port localhost:%d", remotePort, localPort)
		err := sshchannel.Ins().ForwardRemoteToLocal(
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
