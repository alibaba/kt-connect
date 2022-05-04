package transmission

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/service/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

// ForwardPodToLocal mapping pod port to local port
func ForwardPodToLocal(exposePorts, podName, privateKey string) (int, error) {
	log.Info().Msgf("Forwarding pod %s to local via port %s", podName, exposePorts)
	localSshPort := util.GetRandomTcpPort()

	// port forward pod 22 -> local <random port>
	if err := SetupPortForwardToLocal(podName, common.StandardSshPort, localSshPort); err != nil {
		return -1, err
	}

	err := ForwardRemotePortsViaSshTunnel(exposePorts, localSshPort, privateKey)
	if err != nil {
		return -1, err
	}

	return localSshPort, nil
}

// ForwardRemotePortsViaSshTunnel forward multiple remote ports to local
func ForwardRemotePortsViaSshTunnel(exposePorts string, localSshPort int, privateKey string) error {
	// supports multi port-pairs
	portPairs := strings.Split(exposePorts, ",")
	res := make(chan error)
	for _, exposePort := range portPairs {
		localPort, remotePort, err2 := util.ParsePortMapping(exposePort)
		if err2 != nil {
			return err2
		}
		forwardRemotePortViaSshTunnel(localPort, remotePort, localSshPort, privateKey, res)
	}
	select {
	case err := <-res:
		return err
	case <-time.After(1 * time.Second):
		go func() {
			// consume the res channel to avoid block reverse tunnel
			<-res
		}()
	}
	return nil
}

// ForwardRemotePortViaSshTunnel forward remote pod to local
func forwardRemotePortViaSshTunnel(localPort, remotePort, localSshPort int, privateKey string, res chan error) {
	remoteEndpoint := fmt.Sprintf("127.0.0.1:%d", localSshPort)
	localEndpoint := fmt.Sprintf("0.0.0.0:%d", remotePort)
	sshAddress := fmt.Sprintf("127.0.0.1:%d", localPort)
	log.Debug().Msgf("Forwarding %s to local endpoint %s via %s", remoteEndpoint, localEndpoint, sshAddress)
	sshReverseTunnel(privateKey, remoteEndpoint, localEndpoint, sshAddress, res)
}

func sshReverseTunnel(privateKey, remoteEndpoint, localEndpoint, sshAddress string, res chan error) {
	go func() {
		err := sshchannel.Ins().ForwardRemoteToLocal(privateKey, remoteEndpoint, localEndpoint, sshAddress)
		if err != nil {
			if res != nil {
				log.Error().Err(err).Msgf("Failed to setup reverse tunnel")
				res <-err
			} else {
				log.Debug().Err(err).Msgf("Reverse tunnel interrupted")
			}
		}

		time.Sleep(10 * time.Second)
		log.Debug().Msgf("Reverse tunnel reconnecting ...")
		sshReverseTunnel(privateKey, remoteEndpoint, localEndpoint, sshAddress, nil)
	}()
}
