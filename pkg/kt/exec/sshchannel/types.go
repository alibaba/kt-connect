package sshchannel

// Channel network channel
type Channel interface {
	StartSocks5Proxy(privateKey string, sshAddress, socks5Address string) error
	ForwardRemoteToLocal(privateKey string, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(privateKey string, sshAddress, script string) (string, error)
}
