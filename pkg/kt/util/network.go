package util

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

// GetRandomSSHPort get pod random ssh port
func GetRandomSSHPort(podIP string) string {
	return fmt.Sprintf("22%s", podIP[len(podIP)-2:len(podIP)])
}

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (address string) {
	address = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}