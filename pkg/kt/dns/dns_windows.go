package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(k cluster.KubernetesInterface, dnsServer string, opt *options.DaemonOptions) (err error) {
	// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
	if err = util.RunAndWait(exec.Command("netsh",
		"interface",
		"ip",
		"set",
		"dnsservers",
		fmt.Sprintf("name=%s", common.TunNameWin),
		"source=static",
		fmt.Sprintf("address=%s", strings.Split(dnsServer, ":")[0]),
	), opt.Debug); err != nil {
		log.Error().Msgf("Failed to set dns server of tun device")
		return err
	}
	return nil
}

// RestoreDnsServer ...
func (s *Cli) RestoreDnsServer() {
	// Windows dns config is set on device, so explicit removal is unnecessary
}

// GetLocalDomains ...
func GetLocalDomains() string {
	return ""
}

// GetDnsServer get dns server of the default interface
func GetDnsServer() string {
	return ""
}
