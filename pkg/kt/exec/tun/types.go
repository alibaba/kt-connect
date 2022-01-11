package tun

// Tunnel ...
type Tunnel interface {
	CheckContext() error
	ToSocks(sockAddr string, isDebug bool) error
	SetRoute(ipRange []string, isDebug bool) error
	GetName() string
}

// Cli ...
type Cli struct {}
