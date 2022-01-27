package util

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GetRandomSSHPort get pod random ssh port
func GetRandomSSHPort() int {
	for i := 0; i < 10; i++ {
		port := RandomPort()
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			log.Debug().Msgf("port %d not available", port)
			_ = conn.Close()
		} else {
			return port
		}
	}
	return -1
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

// FindBrokenPort Check if all ports has process listening to
// Return empty string if all ports are listened, otherwise return the first broken port
func FindBrokenPort(exposePorts string) string {
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
