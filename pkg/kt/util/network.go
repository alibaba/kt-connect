package util

import (
	"fmt"
	coreV1 "k8s.io/api/core/v1"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GetRandomTcpPort get pod random ssh port
func GetRandomTcpPort() (int, error) {
	for i := 0; i < 20; i++ {
		port := RandomPort()
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			log.Debug().Msgf("Port %d not available", port)
			_ = conn.Close()
		} else {
			log.Debug().Msgf("Using port %d", port)
			return port, nil
		}
	}
	return -1, fmt.Errorf("failed to find an available port")
}

// ParsePortMapping parse <port> or <localPort>:<removePort> parameter
func ParsePortMapping(exposePort string) (int, int, error) {
	localPort := exposePort
	remotePort := exposePort
	ports := strings.SplitN(exposePort, ":", 2)
	if len(ports) > 1 {
		localPort = ports[0]
		remotePort = ports[1]
	}
	lp, err := strconv.Atoi(localPort)
	if err != nil {
		return -1, -1, fmt.Errorf("local port '%s' is not a number", localPort)
	}
	rp, err := strconv.Atoi(remotePort)
	if err != nil {
		return -1, -1, fmt.Errorf("remote port '%s' is not a number", remotePort)
	}
	return lp, rp, nil
}

// WaitPortBeReady return true when port is ready
// It waits at most waitTime seconds, then return false.
func WaitPortBeReady(waitTime, port int) bool {
	for i := 0; i < waitTime; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Debug().Msgf("Waiting for port forward (%s), retry: %d", err, i+1)
			time.Sleep(1 * time.Second)
		} else {
			_ = conn.Close()
			log.Info().Msgf("Port forward connection established")
			return true
		}
	}
	return false
}

// FindBrokenLocalPort Check if all ports has process listening to
// Return empty string if all ports are listened, otherwise return the first broken port
func FindBrokenLocalPort(exposePorts string) string {
	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		localPort := strings.Split(exposePort, ":")[0]
		conn, err := net.Dial("tcp", fmt.Sprintf(":%s", localPort))
		if err == nil {
			_ = conn.Close()
		} else {
			return localPort
		}
	}
	return ""
}

// FindInvalidRemotePort Check if all ports exist in provide service
func FindInvalidRemotePort(exposePorts string, svcPorts []coreV1.ServicePort) string {
	validPorts := make([]string, 0)
	for _, p := range svcPorts {
		validPorts = append(validPorts, strconv.Itoa(p.TargetPort.IntValue()))
	}

	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		splitPorts := strings.Split(exposePort, ":")
		remotePort := splitPorts[0]
		if len(splitPorts) > 1 {
			remotePort = splitPorts[1]
		}
		if !Contains(remotePort, validPorts) {
			return remotePort
		}
	}
	return ""
}

