// +build !windows

package util

import (
	"bufio"
	"os"
	"strings"
)

const (
	ResolvConf       = "/etc/resolv.conf"
	FieldNameserver  = "nameserver"
	FieldDomain      = "domain"
	FieldSearch      = "search"
)

func GetLocalDomains() string {
	f, err := os.Open(ResolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	var localDomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, FieldDomain) {
			localDomains = append(localDomains, strings.TrimSpace(strings.TrimPrefix(line, FieldDomain)))
		} else if strings.HasPrefix(line, FieldSearch) {
			for _, s := range strings.Split(strings.TrimPrefix(line, FieldSearch), " ") {
				if s != "" {
					localDomains = append(localDomains, s)
				}
			}
		}
	}
	return strings.Join(localDomains, ",")
}
