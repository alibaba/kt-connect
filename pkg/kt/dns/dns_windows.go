package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"regexp"
	"strings"
)

// SetNameServer set dns server records
func (s *Cli) SetNameServer(k cluster.KubernetesInterface, dnsServer string, opt *options.DaemonOptions) (err error) {
	// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
	if _, _, err = util.RunAndWait(exec.Command("netsh",
		"interface",
		"ip",
		"set",
		"dnsservers",
		fmt.Sprintf("name=%s", common.TunNameWin),
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
		r, _ := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+")
		return r.FindString(out)
	}
}
