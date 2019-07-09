package util

import (
	"fmt"
	"log"
	"net"
)

// GetOutboundIP Get preferred outbound ip of this machine
func GetOutboundIP() (address string) {
	address = "127.0.0.1"
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	address = fmt.Sprintf("%s", localAddr.IP)
	return
}
