package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(k cluster.KubernetesInterface, dnsServers []string, isDebug bool) (err error) {
	for i, dns := range dnsServers {
		if i == 0 {
			// run command: netsh interface ip set dnsservers name=KtConnectTunnel source=static address=8.8.8.8
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"set",
				"dnsservers",
				fmt.Sprintf("name=%s", common.TunNameWin),
				"source=static",
				fmt.Sprintf("address=%s", strings.Split(dns, ":")[0]),
			), isDebug)
		} else {
			// run command: netsh interface ip add dnsservers name=KtConnectTunnel address=4.4.4.4
			err = util.RunAndWait(exec.Command("netsh",
				"interface",
				"ip",
				"add",
				"dnsservers",
				fmt.Sprintf("name=%s", common.TunNameWin),
				fmt.Sprintf("address=%s", strings.Split(dns, ":")[0]),
			), isDebug)
		}
		if err != nil {
			log.Error().Msgf("Failed to set dns server of tun device")
			return err
		}
	}
	return nil
}

// RestoreDnsServer ...
func (s *Cli) RestoreDnsServer() {
	// Windows dns config is set on device, so explicit removal is unnecessary
}
