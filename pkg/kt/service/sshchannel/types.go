package sshchannel

// Channel network channel
type Channel interface {
	StartSocks5Proxy(privateKey string, sshAddress, socks5Address string) error
	ForwardRemoteToLocal(privateKey string, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(privateKey string, sshAddress, script string) (string, error)
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() Channel {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}
