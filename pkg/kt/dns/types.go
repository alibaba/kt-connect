package dns

import "github.com/alibaba/kt-connect/pkg/kt/cluster"

// DnsConfig ...
type DnsConfig interface {
	SetDnsServer(k cluster.KubernetesInterface, dnsServers []string, isDebug bool) error
	RestoreDnsServer()
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() *Cli {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}