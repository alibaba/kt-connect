package dns

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
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
func (s *Cli) SetNameServer(dnsServer string) error {
	dnsSignal := make(chan error)
	go func() {
		defer func() {
			restoreResolvConf()
			if strings.HasPrefix(opt.Get().Connect.DnsMode, util.DnsModeLocalDns) {
				restoreIptables()
			}
		}()
		if strings.HasPrefix(opt.Get().Connect.DnsMode, util.DnsModeLocalDns) {
			if err := setupIptables(); err != nil {
				dnsSignal <-err
				return
			}
		}
		dnsSignal <- setupResolvConf(dnsServer)

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
	f, err := os.Open(util.ResolvConf)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer

	sample := fmt.Sprintf("%s %s ", util.FieldNameserver, strings.Split(dnsServer, ":")[0])
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, sample) {
			// required dns server already been added
			return nil
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
	nameserverIp := strings.Split(dnsServer, ":")[0]
	buf.WriteString(fmt.Sprintf("%s %s%s\n", util.FieldNameserver, nameserverIp, commentKtAdded))

	stat, _ := f.Stat()
	return ioutil.WriteFile(util.ResolvConf, buf.Bytes(), stat.Mode())
}

func setupIptables() error {
	// run command: iptables --table nat --insert OUTPUT --proto udp --dest 127.0.0.1/32 --dport 53 --jump REDIRECT --to-ports 10053
	if _, _, err := util.RunAndWait(exec.Command("iptables",
		"--table",
		"nat",
		"--insert",
		"OUTPUT",
		"--proto",
		"udp",
		"--dest",
		fmt.Sprintf("%s/32", common.Localhost),
		"--dport",
		strconv.Itoa(common.StandardDnsPort),
		"--jump",
		"REDIRECT",
		"--to-ports",
		strconv.Itoa(util.AlternativeDnsPort),
	)); err != nil {
		log.Error().Msgf("Failed to use local dns server")
		return err
	}
	return nil
}

func restoreResolvConf() {
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

func restoreIptables() {
	for {
		// run command: iptables --table nat --delete OUTPUT --proto udp --dest 127.0.0.1/32 --dport 53 --jump REDIRECT --to-ports 10053
		_, _, err := util.RunAndWait(exec.Command("iptables",
			"--table",
			"nat",
			"--delete",
			"OUTPUT",
			"--proto",
			"udp",
			"--dest",
			fmt.Sprintf("%s/32", common.Localhost),
			"--dport",
			strconv.Itoa(common.StandardDnsPort),
			"--jump",
			"REDIRECT",
			"--to-ports",
			strconv.Itoa(util.AlternativeDnsPort),
		))
		if err != nil {
			// no more rule left
			break
		}
	}
}
