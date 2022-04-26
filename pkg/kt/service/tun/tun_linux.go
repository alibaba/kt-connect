package tun

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() error {
	// TODO: check whether ip command exists
	return nil
}

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	// run command: ip link set dev kt0 up
	_, _, err := util.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.GetName(),
		"up",
	))
	if err != nil {
		log.Error().Msgf("Failed to set tun device up")
		return AllRouteFailError{err}
	}
	var lastErr error
	anyRouteOk := false
	for _, r := range ipRange {
		log.Info().Msgf("Adding route to %s", r)
		// run command: ip route add 10.96.0.0/16 dev kt0
		_, _, err = util.RunAndWait(exec.Command("ip",
			"route",
			"add",
			r,
			"dev",
			s.GetName(),
		))
		if err != nil {
			log.Warn().Msgf("Failed to set route %s to tun device", r)
			lastErr = err
		} else {
			anyRouteOk = true
		}
	}
	if !anyRouteOk {
		return AllRouteFailError{lastErr}
	}
	return lastErr
}

// RestoreRoute delete route rules made by kt
func (s *Cli) RestoreRoute() error {
	return nil
}

func (s *Cli) GetName() string {
	return util.TunNameLinux
}
