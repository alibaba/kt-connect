// +build !windows

package resolvconf

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"strings"
)

const (
	commentKtAdded   = " # added by ktctl"
	commentKtRemoved = " # removed by ktctl"
	resolvConf       = "/etc/resolv.conf"
	fieldNameserver  = "nameserver"
	fieldDomain      = "domain"
	fieldSearch      = "search"
)

func AddNameserver(nameserver string) error {
	f, err := os.Open(resolvConf)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	prefix := fmt.Sprintf("%s %s", fieldNameserver, nameserver)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, fieldNameserver) {
			buf.WriteString("#")
			buf.WriteString(line)
			buf.WriteString(commentKtRemoved)
			buf.WriteString("\n")
		} else if strings.HasPrefix(line, prefix) {
			return nil
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	// Add nameserver and comment to resolv.conf
	buf.WriteString(fmt.Sprintf("%s%s", prefix, commentKtAdded))
	buf.WriteString("\n")

	stat, _ := f.Stat()
	err = ioutil.WriteFile(resolvConf, buf.Bytes(), stat.Mode())
	return err
}

// RestoreConfig remove the nameserver which is added by ktctl.
func RestoreConfig() error {
	f, err := os.Open(resolvConf)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, commentKtRemoved) {
			line = strings.TrimSuffix(line, commentKtRemoved)
			line = strings.TrimPrefix(line, "#")
			buf.WriteString(line)
			buf.WriteString("\n")
		} else if strings.HasSuffix(line, commentKtAdded) {
			log.Info().Msgf("remove line: %s ", line)
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	stat, _ := f.Stat()
	err = ioutil.WriteFile(resolvConf, buf.Bytes(), stat.Mode())
	return err
}

func GetLocalDomains() string {
	f, err := os.Open(resolvConf)
	if err != nil {
		return ""
	}
	defer f.Close()

	var localDomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, fieldDomain) {
			localDomains = append(localDomains, strings.TrimSpace(strings.TrimPrefix(line, fieldDomain)))
		} else if strings.HasPrefix(line, fieldSearch) {
			for _, s := range strings.Split(strings.TrimPrefix(line, fieldSearch), " ") {
				if s != "" {
					localDomains = append(localDomains, s)
				}
			}
		}
	}
	return strings.Join(localDomains, ",")
}
