package sshchannel

// Channel network channel
type Channel interface {
	StartSocks5Proxy(privateKey, sshAddress, socks5Address string) error
	ForwardRemoteToLocal(privateKey, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(privateKey, sshAddress, script string) (string, error)
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
