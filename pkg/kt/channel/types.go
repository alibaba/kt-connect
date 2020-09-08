package channel

// Certificate certificate
type Certificate struct {
	Username   string
	Password   string
	PrivateKey string
}

// Channel network channel
type Channel interface {
	StartSocks5Proxy(certificate *Certificate, sshAddress string, socks5Address string) error
	ForwardRemoteToLocal(certificate *Certificate, sshAddress string, remoteEndpoint string, localEndpoint string) error
}
