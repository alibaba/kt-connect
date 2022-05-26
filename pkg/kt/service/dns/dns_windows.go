package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"strings"
)

// SetNameServer set dns server records
func (s *Cli) SetNameServer(dnsServer string) (err error) {
	// run command: netsh interface ip set interface KtConnectTunnel metric=2
	if _, _, err = util.RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"set",
		"interface",
		util.TunNameWin,
		"metric=2",
	)); err != nil {
		log.Error().Msgf("Failed to set tun device order")
		return err
	}
	// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
	if _, _, err = util.RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"set",
		"dnsservers",
		fmt.Sprintf("name=%s", util.TunNameWin),
		"source=static",
		fmt.Sprintf("address=%s", strings.Split(dnsServer, ":")[0]),
	)); err != nil {
		log.Error().Msgf("Failed to set dns server of tun device")
		return err
	}
	return nil
}

// RestoreNameServer ...
func (s *Cli) RestoreNameServer() {
	// Windows dns config is set on device, so explicit removal is unnecessary
}

// GetLocalDomains ...
func GetLocalDomains() string {
	return ""
}

// GetNameServer get dns server of the default interface
func GetNameServer() string {
	// run command: netsh interface ip show dnsservers
	out, _, err := util.RunAndWait(exec.Command("netsh",
		"interface",
		"ipv4",
		"show",
		"dnsservers",
	))
	if err != nil {
		log.Error().Msgf("Failed to get upstream dns server")
		return ""
	}
	_, _ = util.BackgroundLogger.Write([]byte(">> Get dns: " + out + util.Eol))

	r, _ := regexp.Compile(util.IpAddrPattern)
	nsAddresses := r.FindAllString(out, 10)
	if nsAddresses == nil {
		log.Warn().Msgf("No upstream dns server available")
		return ""
	}
	for _, addr := range nsAddresses {
		if addr != common.Localhost {
			return addr
		}
	}
	log.Warn().Msgf("No valid upstream dns server available")
	return ""
}
