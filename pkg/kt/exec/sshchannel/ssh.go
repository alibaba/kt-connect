package sshchannel

import (
	"io"
	"net"

	"github.com/armon/go-socks5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
)

// SSHChannel ssh channel
type SSHChannel struct{}

// StartSocks5Proxy start socks5 proxy
func (c *SSHChannel) StartSocks5Proxy(certificate *Certificate, sshAddress, socks5Address string) (err error) {
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(conf)
	if err != nil {
		return err
	}

	// Process will hang at here
	if err = serverSocks.ListenAndServe("tcp", socks5Address); err != nil {
		log.Error().Msgf("Failed to create socks5 server", err)
	}
	return
}

// ForwardRemoteToLocal forward remote request to local
func (c *SSHChannel) ForwardRemoteToLocal(certificate *Certificate, sshAddress, remoteEndpoint, localEndpoint string) (err error) {
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	if err != nil {
		log.Error().Msgf("Fail to create ssh tunnel: %s", err)
		return err
	}
	defer conn.Close()

	// Listen on remote server port, process will hang at here
	listener, err := conn.Listen("tcp", remoteEndpoint)
	if err != nil {
		log.Error().Msgf("Fail to listen remote endpoint: %s", err)
		return err
	}
	defer listener.Close()

	log.Info().Msgf("Forward %s to localEndpoint %s", remoteEndpoint, localEndpoint)

	go c.handleConnections(localEndpoint, listener)
	return nil
}

// handleConnections handle incoming connections on reverse forwarded tunnel
func (c *SSHChannel) handleConnections(localEndpoint string, listener net.Listener) {
	for {
		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		local, err := net.Dial("tcp", localEndpoint)
		if err != nil {
			log.Error().Msgf("Dial to local service error: %s", err)
			return
		}

		client, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("Error: %s", err)
			return
		}

		handleClient(client, local)
	}
}

func connection(username string, password string, address string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{
		ssh.Password(password),
	}

	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Error().Msgf("Fail create ssh connection: %s", err)
	}
	return conn, err
}

func handleClient(client net.Conn, remote net.Conn) {
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Error().Msgf("Error while copy remote->local: %s", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Error().Msgf("Error while copy local->remote: %s", err)
		}
		chDone <- true
	}()

	<-chDone
}
