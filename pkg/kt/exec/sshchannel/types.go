package sshchannel

// Certificate certificate
type Certificate struct {
	Username   string
	Password   string
	PrivateKey string
}

// Channel network channel
type Channel interface {
	StartSocks5Proxy(certificate *Certificate, sshAddress, socks5Address string) error
	ForwardRemoteToLocal(certificate *Certificate, sshAddress, remoteEndpoint, localEndpoint string) error
	RunScript(certificate *Certificate, sshAddress, script string) (string, error)
}
