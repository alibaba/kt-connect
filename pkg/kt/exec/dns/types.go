package dns

import "github.com/alibaba/kt-connect/pkg/kt/cluster"

// DnsConfig ...
type DnsConfig interface {
	SetDnsServer(k cluster.KubernetesInterface, dnsServers []string, isDebug bool) error
	RestoreDnsServer()
}

// Cli ...
type Cli struct {}
