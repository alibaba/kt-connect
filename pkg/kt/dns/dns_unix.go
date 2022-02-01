//go:build !windows

package dns

import (
	"bufio"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

// listen address of systemd-resolved
const resolvedAddr = "127.0.0.53"
const resolvedConf = "/run/systemd/resolve/resolv.conf"

// GetLocalDomains get domain search postfixes
func GetLocalDomains() string {
	f, err := os.Open(common.ResolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	var localDomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, common.FieldDomain) {
			localDomains = append(localDomains, strings.TrimSpace(strings.TrimPrefix(line, common.FieldDomain)))
		} else if strings.HasPrefix(line, common.FieldSearch) {
			for _, s := range strings.Split(strings.TrimPrefix(line, common.FieldSearch), " ") {
				if s != "" {
					localDomains = append(localDomains, s)
				}
			}
		}
	}
	return strings.Join(localDomains, ",")
}

// GetNameServer get primary dns server
func GetNameServer() string {
	ns := fetchNameServerInConf(common.ResolvConf)
	if util.IsLinux() && ns == resolvedAddr {
		log.Debug().Msgf("Using systemd-resolved")
		return fetchNameServerInConf(resolvedConf)
	}
	return ns
}

func fetchNameServerInConf(resolvConf string) string {
	f, err := os.Open(resolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, common.FieldNameserver) {
			return strings.TrimSpace(strings.TrimPrefix(line, common.FieldNameserver))
		}
	}
	return ""
}
