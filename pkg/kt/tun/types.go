package tun

// Tunnel ...
type Tunnel interface {
	CheckContext() error
	ToSocks(sockAddr string, isDebug bool) error
	SetRoute(ipRange []string, isDebug bool) error
	GetName() string
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