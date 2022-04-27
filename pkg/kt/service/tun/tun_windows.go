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
		log.Info().Msgf("Adding route to %s", r)
		_, mask, err := toIpAndMask(r)
		tunIp := strings.Split(r, "/")[0]
		if err != nil {
			return AllRouteFailError{err}
		}
		if i == 0 {
			// run command: netsh interface ipv4 set address KtConnectTunnel static 172.20.0.1 255.255.0.0
			_, _, err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ipv4",
				"set",
				"address",
				s.GetName(),
				"static",
				tunIp,
				mask,
			))
		} else {
			// run command: netsh interface ipv4 add address KtConnectTunnel 172.21.0.1 255.255.0.0
			_, _, err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ipv4",
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

// RestoreRoute delete route rules made by kt
func (s *Cli) RestoreRoute() error {
	var lastErr error
	// run command: netsh interface ipv4 show route store=persistent
	out, _, err := util.RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"show",
		"route",
		"store=persistent",
	))
	if err != nil {
		log.Warn().Msgf("failed to get route table")
		return err
	}
	for _, line := range strings.Split(out, util.Eol) {
		// Assume only kt using gateway address of x.x.x.0
		if !strings.HasSuffix(line, ".0") {
			continue
		}
		log.Debug().Msgf("Route recode: %s", line)
		parts := strings.Split(line, " ")
		ipRange := ""
		iface := ""
		gateway := ""
		index := 0
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				index++
				if index == 3 {
					ipRange = parts[i]
					break
				} else if index == 2 {
					iface = parts[i]
				} else if index == 1 {
					gateway = parts[i]
				}
			}
		}
		if ipRange == "" {
			continue
		}
		// run command: netsh interface ipv4 delete route store=persistent 172.20.0.0/16 29 172.20.0.0
		_, _, err = util.RunAndWait(exec.Command("netsh",
			"interface",
			"ipv4",
			"delete",
			"route",
			"store=persistent",
			ipRange,
			iface,
			gateway,
		))
		if err != nil {
			log.Warn().Msgf("Failed to clean route to %s", ipRange)
			lastErr = err
		} else {
			log.Info().Msgf(" * %s", ipRange)
		}
	}
	return lastErr
}

func (s *Cli) GetName() string {
	return util.TunNameWin
}
