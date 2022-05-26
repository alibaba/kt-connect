package tun

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() error {
	if !util.CanRun(exec.Command("which", "ifconfig")) {
		return fmt.Errorf("failed to found 'ifconfig' command")
	}
	if !util.CanRun(exec.Command("which", "route")) {
		return fmt.Errorf("failed to found 'route' command")
	}
	if !util.CanRun(exec.Command("which", "netstat")) {
		return fmt.Errorf("failed to found 'netstat' command")
	}
	return nil
}

// SetRoute set specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	var err, lastErr error
	anyRouteOk := false
	for i, r := range ipRange {
		log.Info().Msgf("Adding route to %s", r)
		tunIp := strings.Split(r, "/")[0]
		if i == 0 {
			// run command: ifconfig utun6 inet 172.20.0.0/16 172.20.0.0
			_, _, err = util.RunAndWait(exec.Command("ifconfig",
				s.GetName(),
				"inet",
				r,
				tunIp,
			))
		} else {
			// run command: ifconfig utun6 add 172.20.0.0/16 172.20.0.1
			_, _, err = util.RunAndWait(exec.Command("ifconfig",
				s.GetName(),
				"add",
				r,
				tunIp,
			))
		}
		if err != nil {
			log.Warn().Msgf("Failed to add ip addr %s to tun device", tunIp)
			lastErr = err
			continue
		} else {
			anyRouteOk = true
		}
		// run command: route add -net 172.20.0.0/16 -interface utun6
		_, _, err = util.RunAndWait(exec.Command("route",
			"add",
			"-net",
			r,
			"-interface",
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

// CheckRoute check whether all route rule setup properly
func (s *Cli) CheckRoute(ipRange []string) []string {
	var failedIpRange []string
	// run command: netstat -rn
	out, _, err := util.RunAndWait(exec.Command("netstat",
		"-rn",
	))
	if err != nil {
		log.Warn().Msgf("Failed to get route table")
		return []string{}
	}
	_, _ = util.BackgroundLogger.Write([]byte(">> Get route: " + out + util.Eol))

	for _, ir := range ipRange {
		found := false
		for _, line := range strings.Split(out, util.Eol) {
			ip := strings.Split(ir, "/")[0]
			if strings.HasPrefix(line, ip) && strings.HasSuffix(strings.TrimSpace(line), s.GetName()) {
				found = true
				break
			}
		}
		if !found {
			failedIpRange = append(failedIpRange, ir)
		}
	}
	return failedIpRange
}

// RestoreRoute delete route rules made by kt
func (s *Cli) RestoreRoute() error {
	// Route will be auto removed when tun device destroyed
	return nil
}

var tunName = ""
func (s *Cli) GetName() string {
	if tunName != "" {
		return tunName
	}
	tunName = fmt.Sprintf("%s%d", util.TunNameMac, 9)
	if ifaces, err := net.Interfaces(); err == nil {
		tunN := 0
		for _, i := range ifaces {
			if strings.HasPrefix(i.Name, util.TunNameMac) {
				if num, err2 := strconv.Atoi(strings.TrimPrefix(i.Name, util.TunNameMac)); err2 == nil && num > tunN {
					tunN = num
				}
			}
		}
		tunName = fmt.Sprintf("%s%d", util.TunNameMac, tunN + 1)
	}
	return tunName
}
