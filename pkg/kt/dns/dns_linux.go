package dns

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	commentKtAdded   = " # Added by KtConnect"
	commentKtRemoved = " # Removed by KtConnect"
)

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(k cluster.KubernetesInterface, dnsServers []string, isDebug bool) error {
	dnsSignal := make(chan error)
	go func() {
		f, err := os.Open(util.ResolvConf)
		if err != nil {
			dnsSignal <-err
			return
		}
		defer f.Close()

		var buf bytes.Buffer

		sample := fmt.Sprintf("%s %s ", util.FieldNameserver, dnsServers[0])
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, sample) {
				// required dns server already been added
				dnsSignal <-nil
				return
			} else if strings.HasPrefix(line, util.FieldNameserver) {
				buf.WriteString("#")
				buf.WriteString(line)
				buf.WriteString(commentKtRemoved)
				buf.WriteString("\n")
			} else {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}

		// Add nameserver and comment to resolv.conf
		for _, nameserver := range dnsServers {
			buf.WriteString(fmt.Sprintf("%s %s%s\n", util.FieldNameserver, nameserver, commentKtAdded))
		}

		stat, _ := f.Stat()
		dnsSignal <-ioutil.WriteFile(util.ResolvConf, buf.Bytes(), stat.Mode())

		defer s.RestoreDnsServer()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
	}()
	return <-dnsSignal
}

// RestoreDnsServer remove the nameservers added by ktctl
func (s *Cli) RestoreDnsServer() {
	f, err := os.Open(util.ResolvConf)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open resolve.conf during restoring")
		return
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
			log.Debug().Msgf("remove line: %s ", line)
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	stat, _ := f.Stat()
	if err = ioutil.WriteFile(util.ResolvConf, buf.Bytes(), stat.Mode()); err != nil {
		log.Error().Err(err).Msgf("Failed to write resolve.conf during restoring")
	}
}
