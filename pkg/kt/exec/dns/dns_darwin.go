package dns

import (
	"os"
	"os/signal"
	"syscall"
)

// SetDnsServer set dns server records
func (s *Cli) SetDnsServer(dnsServers []string, isDebug bool) error {
	dnsSignal := make(chan error)
	go func() {
		dnsSignal <-nil

		defer restoreDnsServer()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}()
	return <-dnsSignal
}

// restoreDnsServer remove the nameservers added by ktctl
func restoreDnsServer() {
}
