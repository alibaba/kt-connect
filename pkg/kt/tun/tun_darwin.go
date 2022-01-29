package tun

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

// CheckContext check everything needed for tun setup
func (s *Cli) CheckContext() error {
	// TODO: check whether ifconfig and route command exists
	return nil
}

// SetRoute set specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string, isDebug bool) error {
	var err, lastErr error
	for i, r := range ipRange {
		log.Info().Msgf("Adding route to %s", r)
		tunIp := strings.Split(r, "/")[0]
		if i == 0 {
			// run command: ifconfig utun6 inet 172.20.0.0/16 172.20.0.0
			err = util.RunAndWait(exec.Command("ifconfig",
				s.GetName(),
				"inet",
				r,
				tunIp,
			), isDebug)
		} else {
			// run command: ifconfig utun6 add 172.20.0.0/16 172.20.0.1
			err = util.RunAndWait(exec.Command("ifconfig",
				s.GetName(),
				"add",
				r,
				tunIp,
			), isDebug)
		}
		if err != nil {
			log.Warn().Msgf("Failed to add ip addr %s to tun device", tunIp)
			lastErr = err
			continue
		}
		// run command: route add -net 172.20.0.0/16 -interface utun6
		err = util.RunAndWait(exec.Command("route",
			"add",
			"-net",
			r,
			"-interface",
			s.GetName(),
		), isDebug)
		if err != nil {
			log.Warn().Msgf("Failed to set route %s to tun device", r)
			lastErr = err
		}
	}
	return lastErr
}

var tunName = ""
func (s *Cli) GetName() string {
	if tunName != "" {
		return tunName
	}
	tunName = fmt.Sprintf("%s%d", common.TunNameMac, 9)
	if ifaces, err := net.Interfaces(); err == nil {
		tunN := 0
		for _, i := range ifaces {
			if strings.HasPrefix(i.Name, common.TunNameMac) {
				if num, err2 := strconv.Atoi(strings.TrimPrefix(i.Name, common.TunNameMac)); err2 == nil && num > tunN {
					tunN = num
				}
			}
		}
		tunName = fmt.Sprintf("%s%d", common.TunNameMac, tunN + 1)
	}
	return tunName
}
