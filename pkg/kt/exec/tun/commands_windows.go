package tun

import (
	"fmt"
	ktexec "github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
)

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	for i, r := range ipRange {
		ip, mask, err := toIpAndMask(r)
		tunIp := firstIp(r)
		if err != nil {
			return err
		}
		if i == 0 {
			// run command: netsh interface ip set address KtConnectTunnel static 172.20.0.1 255.255.0.0
			err = ktexec.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"address",
				s.getTunName(),
				"static",
				tunIp,
				mask,
			), "add_ip_addr")
		} else {
			// run command: netsh interface ip add address KtConnectTunnel 172.21.0.1 255.255.0.0
			err = ktexec.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"address",
				s.getTunName(),
				tunIp,
				mask,
			), "add_ip_addr")
		}
		if err != nil {
			log.Error().Err(err).Msgf("Failed to add ip addr %s to tun device", tunIp)
			return err
		}
		// run command: route add 172.20.0.0 mask 255.255.0.0 172.20.0.1
		err = ktexec.RunAndWait(exec.Command("route",
			"add",
			ip,
			"mask",
			mask,
			tunIp,
		), "add_route")
		if err != nil {
			log.Error().Err(err).Msgf("Failed to set route %s to tun device", r)
			return err
		}
	}
	return nil
}

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(dnsServers []string) (err error) {
	for i, dns := range dnsServers {
		if i == 0 {
			// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
			err = ktexec.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"dnsservers",
				fmt.Sprintf("name=%s", s.getTunName()),
				"source=static",
				fmt.Sprintf("address=%s", dns),
			), "add_dns_server")
		} else {
			// run command: netsh interface ip add dnsservers name=KtConnectTunnel address=4.4.4.4
			err = ktexec.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"dnsservers",
				fmt.Sprintf("name=%s", s.getTunName()),
				fmt.Sprintf("address=%s", dns),
			), "add_dns_server")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Cli) getTunName() string {
	return "KtConnectTunnel"
}
