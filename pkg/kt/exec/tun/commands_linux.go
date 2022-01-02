package tun

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

const (
	commentKtAdded   = " # added by ktctl"
	commentKtRemoved = " # removed by ktctl"
) 

// SetRoute let specified ip range route to tun device
func (s *Cli) SetRoute(ipRange []string) error {
	// run command: ip link set dev kt0 up
	err := util.RunAndWait(exec.Command("ip",
		"link",
		"set",
		"dev",
		s.getTunName(),
		"up",
	), "set_device_up")
	if err != nil {
		log.Error().Err(err).Msgf("Failed to set tun device up")
		return err
	}
	for _, r := range ipRange {
		// run command: ip route add 10.96.0.0/16 dev kt0
		err = util.RunAndWait(exec.Command("ip",
			"route",
			"add",
			r,
			"dev",
			s.getTunName(),
		), "add_route")
		if err != nil {
			log.Error().Err(err).Msgf("Failed to set route %s to tun device", r)
			return err
		}
	}
	return nil
}

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(dnsServers []string) error {
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

		defer RestoreDnsServer()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}()
	return <-dnsSignal
}

// RestoreDnsServer remove the nameservers added by ktctl
func RestoreDnsServer() error {
	f, err := os.Open(util.ResolvConf)
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
			log.Debug().Msgf("remove line: %s ", line)
		} else {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	stat, _ := f.Stat()
	err = ioutil.WriteFile(util.ResolvConf, buf.Bytes(), stat.Mode())
	return err
}

func (s *Cli) getTunName() string {
	return "kt0"
}
