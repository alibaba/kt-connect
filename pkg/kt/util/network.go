package util

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GetRandomSSHPort get pod random ssh port
func GetRandomSSHPort(podIP string) string {
	parts := strings.Split(podIP, ".")
	rdm := parts[len(parts)-1]

	if len(rdm) == 1 {
		rdm = fmt.Sprintf("0%s", rdm)
	}

	if len(rdm) > 2 {
		rdm = rdm[len(rdm)-2:]
	}

	return fmt.Sprintf("22%s", rdm)
}

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (address string) {
	address = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}

// WaitPortBeReady return true when port is ready
// It waits at most waitTime seconds, then return false.
func WaitPortBeReady(waitTime, port int) bool {
	for i := 0; i < waitTime; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Debug().Msgf("Connect to port-forward failed, error: %s, retry: %d", err, i)
			time.Sleep(1 * time.Second)
		} else {
			conn.Close()
			log.Info().Msgf("Connect to port-forward successful")
			return true
		}
	}
	return false
}

// ExtractNetMaskFromCidr extract net mask length (e.g. 16) from cidr (e.g. 1.2.3.4/16)
func ExtractNetMaskFromCidr(cidr string) string {
	return cidr[strings.Index(cidr, "/")+1:]
}
