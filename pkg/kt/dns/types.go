package dns

// DnsConfig ...
type DnsConfig interface {
	SetNameServer(dnsServers []string, isDebug bool) error
	RestoreNameServer()
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
