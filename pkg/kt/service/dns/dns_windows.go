package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"strings"
)

// SetNameServer set dns server records
func (s *Cli) SetNameServer(dnsServer string) (err error) {
	// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
	if _, _, err = util.RunAndWait(exec.Command("netsh",
		"interface",
		"ip",
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
	if out, _, err := util.RunAndWait(exec.Command("netsh",
		"interface",
		"ip",
		"show",
		"dnsservers",
	)); err != nil {
		log.Error().Msgf("Failed to get dns server")
		return ""
	} else {
		r, _ := regexp.Compile(util.IpAddrPattern)
		return r.FindString(out)
	}
}
