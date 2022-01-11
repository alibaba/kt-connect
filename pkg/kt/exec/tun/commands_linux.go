package tun

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
)

const (
	commentKtAdded   = " # added by ktctl"
	commentKtRemoved = " # removed by ktctl"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() error {
	// TODO: check whether ip command exists
	return nil
}

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string, isDebug bool) error {
	// run command: ip link set dev kt0 up
	err := util.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.GetName(),
		"up",
	), isDebug)
	if err != nil {
		log.Error().Msgf("Failed to set tun device up")
		return err
	}
	var lastErr error
	for _, r := range ipRange {
		// run command: ip route add 10.96.0.0/16 dev kt0
		err = util.RunAndWait(exec.Command("ip",
			"route",
			"add",
			r,
			"dev",
			s.GetName(),
		), isDebug)
		if err != nil {
			log.Warn().Msgf("Failed to set route %s to tun device", r)
			lastErr = err
		}
	}
	return lastErr
}

func (s *Cli) GetName() string {
	return common.TunNameLinux
}
