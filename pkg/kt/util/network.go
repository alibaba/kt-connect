package util

import (
	"fmt"
	"net"
	"strings"

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
		log.Error().Err(err).Send()
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}

// ExtractNetMaskFromCidr extract net mask length (e.g. 16) from cidr (e.g. 1.2.3.4/16)
func ExtractNetMaskFromCidr(cidr string) string {
	return cidr[strings.Index(cidr, "/")+1:]
}
