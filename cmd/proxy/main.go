package main

import (
	"github.com/alibaba/kt-connect/pkg/proxy/daemon"
)

func main() {
	daemon.SSHDStart()
	daemon.DNSServerStart()
}
