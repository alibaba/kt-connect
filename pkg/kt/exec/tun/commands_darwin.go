package tun

import (
	ktexec "github.com/alibaba/kt-connect/pkg/kt/util"
	"os/exec"
)

// SetRoute set specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	for _, r := range ipRange {
		// run command: ifconfig utun6 inet 172.20.0.0/16 172.20.0.1
		err := ktexec.RunAndWait(exec.Command("ifconfig",
			s.TunName,
			"inet",
			r,
			firstIp(r),
		), "add_ip_addr")
		if err != nil {
			return err
		}
		// run command: route add -net 172.20.0.0/16 -iface 172.20.0.1
		ktexec.RunAndWait(exec.Command("route",
			"add",
			"-net",
			r,
			"-iface",
			firstIp(r),
		), "add_route")
		if err != nil {
			return err
		}
	}
	return nil
}

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(dnsServers []string) error {
	return nil
}

func (s *Cli) getTunName() string {
	return "utun6"
}
