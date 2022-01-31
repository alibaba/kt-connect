//go:build !windows

package dns

import (
	"bufio"
	"github.com/alibaba/kt-connect/pkg/common"
	"os"
	"strings"
)

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

// GetDnsServer get primary dns server
func GetDnsServer() string {
	f, err := os.Open(common.ResolvConf)
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
