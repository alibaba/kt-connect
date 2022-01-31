package dns

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	commentKtAdded   = " # Added by KtConnect"
	commentKtRemoved = " # Removed by KtConnect"
)

// SetNameServer set dns server records
func (s *Cli) SetNameServer(k cluster.KubernetesInterface, dnsServer string, opt *options.DaemonOptions) error {
	dnsSignal := make(chan error)
	go func() {
		if opt.ConnectOptions.DnsMode == common.DnsModeLocalDns {
			dnsSignal <-setupIptables(opt)
		} else {
			dnsSignal <-setupResolvConf(dnsServer)
		}

		defer func() {
			if opt.ConnectOptions.DnsMode == common.DnsModeLocalDns {
				restoreIptables()
			} else {
				restoreResolvConf()
			}
		}()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
	}()
	return <-dnsSignal
}

// RestoreNameServer remove the nameservers added by ktctl
func (s *Cli) RestoreNameServer() {
	restoreResolvConf()
	restoreIptables()
}

func setupResolvConf(dnsServer string) error {
	f, err := os.Open(common.ResolvConf)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	sample := fmt.Sprintf("%s %s ", common.FieldNameserver, strings.Split(dnsServer, ":")[0])
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, sample) {
			// required dns server already been added
			return nil
		} else if strings.HasPrefix(line, common.FieldNameserver) {
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
	nameserverIp := strings.Split(dnsServer, ":")[0]
	buf.WriteString(fmt.Sprintf("%s %s%s\n", common.FieldNameserver, nameserverIp, commentKtAdded))

	stat, _ := f.Stat()
	return ioutil.WriteFile(common.ResolvConf, buf.Bytes(), stat.Mode())
}

func setupIptables(opt *options.DaemonOptions) error {
	// run command: iptables --table nat --insert OUTPUT --proto udp --dport 53 --jump REDIRECT --to-ports 10053
	if _, _, err := util.RunAndWait(exec.Command("iptables",
		"--table",
		"nat",
		"--insert",
		"OUTPUT",
		"--proto",
		"udp",
		"--dport",
		strconv.Itoa(common.StandardDnsPort),
		"--jump",
		"REDIRECT",
		"--to-ports",
		strconv.Itoa(common.AlternativeDnsPort),
	)); err != nil {
		log.Error().Msgf("Failed to use local dns server")
		return err
	}
	return nil
}

func restoreResolvConf() {
	f, err := os.Open(common.ResolvConf)
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
	if err = ioutil.WriteFile(common.ResolvConf, buf.Bytes(), stat.Mode()); err != nil {
		log.Error().Err(err).Msgf("Failed to write resolve.conf during restoring")
	}
}

func restoreIptables() {
	for {
		// run command: iptables --table nat --delete OUTPUT --proto udp --dport 53 --jump REDIRECT --to-ports 10053
		_, _, err := util.RunAndWait(exec.Command("iptables",
			"--table",
			"nat",
			"--delete",
			"OUTPUT",
			"--proto",
			"udp",
			"--dport",
			strconv.Itoa(common.StandardDnsPort),
			"--jump",
			"REDIRECT",
			"--to-ports",
			strconv.Itoa(common.AlternativeDnsPort),
		))
		if err != nil {
			// no more rule left
			break
		}
	}
}
