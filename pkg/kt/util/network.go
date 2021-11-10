package util

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GetRandomSSHPort get pod random ssh port
func GetRandomSSHPort() int {
	for i := 0; i < 10; i++ {
		port := rand.Intn(65535-1024) + 1024
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = conn.Close()
			return port
		}
	}
	return -1
}

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (address string) {
	address = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}

// GetLocalIps Get all local ip addresses
func GetLocalIps() (ips []string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.To4().String())
				}
			}
		}
	}
	return
}

// ExtractNetMaskFromCidr extract net mask length (e.g. 16) from cidr (e.g. 1.2.3.4/16)
func ExtractNetMaskFromCidr(cidr string) string {
	return cidr[strings.Index(cidr, "/")+1:]
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
