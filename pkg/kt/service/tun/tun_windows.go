package tun

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	wintun "golang.zx2c4.com/wintun"
	"os/exec"
	"strings"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to found tun driver: %v", r)
		}
	}()
	wintun.RunningVersion()
	return
}

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	var lastErr error
	anyRouteOk := false
	for i, r := range ipRange {
		_, mask, err := toIpAndMask(r)
		tunIp := strings.Split(r, "/")[0]
		if err != nil {
			return AllRouteFailError{err}
		}
		if i == 0 {
			// run command: netsh interface ip set address KtConnectTunnel static 172.20.0.1 255.255.0.0
			_, _, err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"address",
				s.GetName(),
				"static",
				tunIp,
				mask,
			))
		} else {
			// run command: netsh interface ip add address KtConnectTunnel 172.21.0.1 255.255.0.0
			_, _, err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"address",
				s.GetName(),
				tunIp,
				mask,
			))
		}
		if err != nil {
			log.Warn().Msgf("Failed to add ip addr %s to tun device", tunIp)
			lastErr = err
			continue
		} else {
			anyRouteOk = true
		}
		// run command: netsh interface ipv4 add route 172.20.0.0/16 KtConnectTunnel 172.20.0.0
		_, _, err = util.RunAndWait(exec.Command("netsh",
			"interface",
			"ipv4",
			"add",
			"route",
			r,
			s.GetName(),
			tunIp,
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

func (s *Cli) GetName() string {
	return util.TunNameWin
}
