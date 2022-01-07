package tun

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	var lastErr error
	for i, r := range ipRange {
		ip, mask, err := toIpAndMask(r)
		tunIp := strings.Split(r, "/")[0]
		if err != nil {
			return err
		}
		if i == 0 {
			// run command: netsh interface ip set address KtConnectTunnel static 172.20.0.1 255.255.0.0
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"address",
				s.GetName(),
				"static",
				tunIp,
				mask,
			), "add_ip_addr")
		} else {
			// run command: netsh interface ip add address KtConnectTunnel 172.21.0.1 255.255.0.0
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"address",
				s.GetName(),
				tunIp,
				mask,
			), "add_ip_addr")
		}
		if err != nil {
			log.Warn().Msgf("Failed to add ip addr %s to tun device", tunIp)
			lastErr = err
			continue
		}
		// run command: route add 172.20.0.0 mask 255.255.0.0 172.20.0.1
		err = util.RunAndWait(exec.Command("route",
			"add",
			ip,
			"mask",
			mask,
			tunIp,
		), "add_route")
		if err != nil {
			log.Warn().Msgf("Failed to set route %s to tun device", r)
			lastErr = err
		}
	}
	return lastErr
}

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(dnsServers []string) (err error) {
	// Windows dns config is set on device, so explicit removal is unnecessary
	for i, dns := range dnsServers {
		if i == 0 {
			// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"dnsservers",
				fmt.Sprintf("name=%s", s.GetName()),
				"source=static",
				fmt.Sprintf("address=%s", dns),
			), "add_dns_server")
		} else {
			// run command: netsh interface ip add dnsservers name=KtConnectTunnel address=4.4.4.4
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"dnsservers",
				fmt.Sprintf("name=%s", s.GetName()),
				fmt.Sprintf("address=%s", dns),
			), "add_dns_server")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Cli) GetName() string {
	return "KtConnectTunnel"
}
