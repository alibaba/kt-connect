package tun

import (
	ktexec "github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
)

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	// run command: ip link set dev kt0 up
	err := ktexec.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.getTunName(),
		"up",
	), "set_device_up")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set tun device up")
		return err
	}
	for _, r := range ipRange {
		// run command: ip route add 10.96.0.0/16 dev kt0
		err = ktexec.RunAndWait(exec.Command("ip",
			"route",
			"add",
			r,
			"dev",
			s.getTunName(),
		), "add_route")
		if err != nil {
			log.Error().Err(err).Msgf("Failed to set route %s to tun device", r)
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
	return "kt0"
}
