package tun

// CliInterface ...
type CliInterface interface {
	ToSocks(sockAddr string) error
	SetRoute(ipRange []string) error
	SetDnsServer(dnsServers []string) error
	AddDevice() error
	AddRoute(cidr string) error
	SetDeviceIP() error
	RemoveDevice() error
}

// Cli ...
type Cli struct {
	TunName  string
	SourceIP string
	DestIP   string
	MaskLen  string
}
