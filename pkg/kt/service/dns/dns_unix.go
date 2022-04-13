//go:build !windows

package dns

import (
	"bufio"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
	"strings"
)

// listen address of systemd-resolved
const resolvedAddr = "127.0.0.53"
const resolvedConf = "/run/systemd/resolve/resolv.conf"

// GetLocalDomains get domain search postfixes
func GetLocalDomains() string {
	f, err := os.Open(util.ResolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	var localDomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, util.FieldDomain) {
			localDomains = append(localDomains, strings.TrimSpace(strings.TrimPrefix(line, util.FieldDomain)))
		} else if strings.HasPrefix(line, util.FieldSearch) {
			for _, s := range strings.Split(strings.TrimPrefix(line, util.FieldSearch), " ") {
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
	ns := fetchNameServerInConf(util.ResolvConf)
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
	pattern, _ := regexp.Compile(fmt.Sprintf("^%s[ \t]+" + util.IpAddrPattern, util.FieldNameserver))
	for scanner.Scan() {
		line := scanner.Text()
		if pattern.MatchString(line) {
			return strings.TrimSpace(strings.TrimPrefix(line, util.FieldNameserver))
		}
	}
	return ""
}
