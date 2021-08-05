package sshtunnelling

import (
	"fmt"
	"os/exec"
)

// AddRoute add route to kubernetes network.
func (s *Cli) AddRoute(cidr string) *exec.Cmd {
	// run command: ip route add 10.96.0.0/16 dev tun0
	cmd := exec.Command("ip",
		"route",
		"add",
		cidr,
		"dev",
		s.TunName,
	)
	return cmd
}

// AddDevice add a tun device on machine
func (s *Cli) AddDevice() *exec.Cmd {
	// run command: ip tuntap add dev tun0 mod tun
	return exec.Command("ip",
		"tuntap",
		"add",
		"dev",
		s.TunName,
		"mod",
		"tun",
	)
}

// SetDeviceIP set the ip of tun device
func (s *Cli) SetDeviceIP() *exec.Cmd {
	// run command: ip address add 10.1.1.1/30 dev tun0
	return exec.Command("ip",
		"address",
		"add",
		fmt.Sprintf("%s/%s", s.SourceIP, s.MaskLen),
		"dev",
		s.TunName,
	)
}

func (s *Cli) SetDeviceUp() *exec.Cmd {
	// run command: ip link set dev tun0 up
	return exec.Command("ip",
		"link",
		"set",
		"dev",
		s.TunName,
		"up",
	)
}

func (s *Cli) RemoveDevice() *exec.Cmd {
	// run command: ip link delete tun0
	return exec.Command("ip",
		"link",
		"delete",
		s.TunName,
	)
}
