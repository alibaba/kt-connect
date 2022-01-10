package tun

// CliInterface ...
type CliInterface interface {
	CheckContext() error
	ToSocks(sockAddr string, isDebug bool) error
	SetRoute(ipRange []string, isDebug bool) error
	SetDnsServer(dnsServers []string, isDebug bool) error
	GetName() string
}

// Cli ...
type Cli struct {}
