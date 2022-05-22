package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

const IpAddrPattern = "[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+"

// GetRandomTcpPort get pod random ssh port
func GetRandomTcpPort() int {
	for i := 0; i < 20; i++ {
		port := RandomPort()
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			log.Debug().Msgf("Port %d not available", port)
			_ = conn.Close()
		} else {
			log.Debug().Msgf("Using port %d", port)
			return port
		}
	}
	port := RandomPort()
	log.Info().Msgf("Using random port %d", port)
	return port
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
func FindInvalidRemotePort(exposePorts string, svcPorts map[int]string) string {
	validPorts := make([]string, 0)
	for p := range svcPorts {
		validPorts = append(validPorts, strconv.Itoa(p))
	}
	log.Debug().Msgf("Service target ports: %v", validPorts)

	portPairs := strings.Split(exposePorts, ",")
	for _, exposePort := range portPairs {
		splitPorts := strings.Split(exposePort, ":")
		remotePort := splitPorts[0]
		if len(splitPorts) > 1 {
			remotePort = splitPorts[1]
		}
		if !Contains(validPorts, remotePort) {
			return remotePort
		}
	}
	return ""
}

// ExtractHostIp Get host ip address from url
func ExtractHostIp(url string) string {
	if !strings.Contains(url, ":") {
		return ""
	}
	host := strings.Trim(strings.Split(url, ":")[1], "/")
	if ok, err := regexp.MatchString(IpAddrPattern, host); ok && err == nil {
		return host
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	for _, ip := range ips {
		// skip ipv6
		if ok, err2 := regexp.MatchString(IpAddrPattern, ip.String()); ok && err2 == nil {
			return ip.String()
		}
	}
	return ""
}
