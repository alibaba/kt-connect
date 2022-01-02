package tun

// CliInterface ...
type CliInterface interface {
	ToSocks(sockAddr string) error
	SetRoute(ipRange []string) error
	SetDnsServer(dnsServers []string) error
}

// Cli ...
type Cli struct {}
