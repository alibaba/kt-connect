package dns

// DnsConfig ...
type DnsConfig interface {
	SetNameServer(dnsServer string) error
	RestoreNameServer()
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() DnsConfig {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}
