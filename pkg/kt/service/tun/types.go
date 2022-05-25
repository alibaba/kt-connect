package tun

// Tunnel ...
type Tunnel interface {
	CheckContext() error
	ToSocks(sockAddr string) error
	SetRoute(ipRange []string) error
	CheckRoute(ipRange []string) []string
	RestoreRoute() error
	GetName() string
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() Tunnel {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}