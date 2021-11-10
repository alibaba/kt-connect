package connect

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Inbound mapping local port from cluster
func (s *Shadow) Inbound(exposePorts, podName string) error {
	log.Info().Msgf("Forwarding pod %s to local via port %s", podName, exposePorts)
	localSSHPort := util.GetRandomSSHPort()
	if localSSHPort < 0 {
		return fmt.Errorf("failed to find any available local port")
	}

	// port forward pod 22 -> local <random port>
	cli := exec.Cli{}
	_, _, err := forwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, localSSHPort)
	if err != nil {
		return err
	}

	if s.Options.ExchangeOptions != nil && s.Options.ExchangeOptions.Method == common.ExchangeMethodEphemeral {
		err = exchangeWithEphemeralContainer(exposePorts, localSSHPort)
		if err != nil {
			return err
		}
	} else {
		// remote forward pod -> local via ssh
		exposeLocalPorts(exposePorts, localSSHPort)
	}
	return nil
}

func remoteRedirectPort(exposePorts string, listenedPorts map[string]struct{}) (redirectPort map[string]string, err error) {
	portPairs := strings.Split(exposePorts, ",")
	redirectPort = make(map[string]string)
	for _, exposePort := range portPairs {
		_, remotePort := getPortMapping(exposePort)
		port := randPort(listenedPorts)
		if port == "" {
			return nil, fmt.Errorf("failed to find redirect port for port: %s", remotePort)
		}
		redirectPort[remotePort] = port
	}

	return redirectPort, nil
}

func exchangeWithEphemeralContainer(exposePorts string, localSSHPort int) error {
	ssh := sshchannel.SSHChannel{}
	// Get all listened ports on remote host
	listenedPorts, err := getListenedPorts(&ssh, localSSHPort)
	if err != nil {
		return err
	}

	redirectPorts, err := remoteRedirectPort(exposePorts, listenedPorts)
	if err != nil {
		return err
	}
	var redirectPortStr string
	for k, v := range redirectPorts {
		redirectPortStr += fmt.Sprintf("%s:%s,", k, v)
	}
	redirectPortStr = redirectPortStr[:len(redirectPortStr)-1]
	err = setupIptables(&ssh, redirectPortStr, localSSHPort)
	if err != nil {
		return err
	}
	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		localPort, remotePort := getPortMapping(exposePort)
		var wg sync.WaitGroup
		exposeLocalPort(&wg, &ssh, localPort, redirectPorts[remotePort], localSSHPort)
		wg.Done()
	}

	return nil
}

func randPort(listenedPorts map[string]struct{}) string {
	for i := 0; i < 100; i++ {
		port := strconv.Itoa(rand.Intn(65535-1024) + 1024)
		if _, exists := listenedPorts[port]; !exists {
			return port
		}
	}
	return ""
}

func setupIptables(ssh sshchannel.Channel, redirectPorts string, localSSHPort int) error {
	res, err := ssh.RunScript(
		&sshchannel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", localSSHPort),
		fmt.Sprintf("/setup_iptables.sh %s", redirectPorts))

	if err != nil {
		log.Error().Msgf("Setup iptables failed, error: %s", err)
	}

	log.Debug().Msgf("Run setup iptables result: %s", res)
	return err
}

func getListenedPorts(ssh sshchannel.Channel, localSSHPort int) (map[string]struct{}, error) {
	result, err := ssh.RunScript(
		&sshchannel.Certificate{
			Username: "root",
			Password: "root",
		},
		fmt.Sprintf("127.0.0.1:%d", localSSHPort),
		`netstat -tuln | grep -E '^(tcp|udp|tcp6)' |grep LISTEN |awk '{print $4}' | awk -F: '{printf("%s\n", $NF)}'`)

	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Run get listened ports result: %s", result)
	var listenedPorts = make(map[string]struct{})
	// The result should be a string like
	// 38059
	// 22
	parts := strings.Split(result, "\n")
	for i := range parts {
		if len(parts[i]) > 0 {
			listenedPorts[parts[i]] = struct{}{}
		}
	}

	return listenedPorts, nil
}

func exposeLocalPorts(exposePorts string, localSSHPort int) {
	var wg sync.WaitGroup
	ssh := sshchannel.SSHChannel{}
	// supports multi port pairs
	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		localPort, remotePort := getPortMapping(exposePort)
		exposeLocalPort(&wg, &ssh, localPort, remotePort, localSSHPort)
	}
	wg.Wait()
}

func exposeLocalPort(wg *sync.WaitGroup, ssh sshchannel.Channel, localPort, remotePort string, localSSHPort int) {
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		log.Debug().Msgf("Exposing remote pod:%s to local port localhost:%s", remotePort, localPort)
		err := ssh.ForwardRemoteToLocal(
			&sshchannel.Certificate{
				Username: "root",
				Password: "root",
			},
			fmt.Sprintf("127.0.0.1:%d", localSSHPort),
			fmt.Sprintf("0.0.0.0:%s", fmt.Sprintf("%s", remotePort)),
			fmt.Sprintf("127.0.0.1:%s", localPort),
		)
		if err != nil {
			log.Error().Msgf("Error happen when forward remote request to local %s", err)
		}
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
