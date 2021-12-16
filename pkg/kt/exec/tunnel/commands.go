package tunnel

import (
	"fmt"
	ktexec "github.com/alibaba/kt-connect/pkg/kt/util"
	"os/exec"
)

// AddDevice add a tun device on machine
func (s *Cli) AddDevice() error {
	// run command: ip tuntap add dev tun0 mod tun
	err := ktexec.RunAndWait(exec.Command("ip",
		"tuntap",
		"add",
		"dev",
		s.TunName,
		"mod",
		"tun",
	), "add_device")
	if err != nil {
		return err
	}
	// run command: ip link set dev tun0 up
	return ktexec.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.TunName,
		"up",
	), "set_device_up")
}

// AddRoute add route to kubernetes network.
func (s *Cli) AddRoute(cidr string) error {
	// run command: ip route add 10.96.0.0/16 dev tun0
	return ktexec.RunAndWait(exec.Command("ip",
		"route",
		"add",
		cidr,
		"dev",
		s.TunName,
	), "add_route")
}

// SetDeviceIP set the ip of tun device
func (s *Cli) SetDeviceIP() error {
	// run command: ip address add 10.1.1.1/30 dev tun0
	return ktexec.RunAndWait(exec.Command("ip",
		"address",
		"add",
		fmt.Sprintf("%s/%s", s.SourceIP, s.MaskLen),
		"dev",
		s.TunName,
	), "set_device_ip")
}

func (s *Cli) RemoveDevice() error {
	// run command: ip link delete tun0
	return ktexec.RunAndWait(exec.Command("ip",
		"link",
		"delete",
		s.TunName,
	), "remove device")
}
