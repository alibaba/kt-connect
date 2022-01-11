package dns

// DnsConfig ...
type DnsConfig interface {
	SetDnsServer(dnsServers []string, isDebug bool) error
}

// Cli ...
type Cli struct {}
