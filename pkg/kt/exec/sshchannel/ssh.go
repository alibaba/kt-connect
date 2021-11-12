package sshchannel

import (
	"bytes"
	"io"
	"net"
	"sync"

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
		log.Error().Err(err).Msgf("Failed to create socks5 server")
	}
	return
}

// RunScript run the script on remote host.
func (c *SSHChannel) RunScript(certificate *Certificate, sshAddress, script string) (result string, err error) {
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	if err != nil {
		log.Error().Err(err).Msgf("Fail to create ssh tunnel")
		return "", err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Error().Err(err).Msgf("Fail to create ssh session")
		return "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	err = session.Run(script)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to run ssh script")
		return "", err
	}
	output := stdoutBuf.String()
	return output, nil
}

// ForwardRemoteToLocal forward remote request to local
func (c *SSHChannel) ForwardRemoteToLocal(certificate *Certificate, sshAddress, remoteEndpoint, localEndpoint string) (err error) {
	conn, err := connection(certificate.Username, certificate.Password, sshAddress)
	if err != nil {
		log.Error().Err(err).Msgf("Fail to create ssh tunnel")
		return err
	}
	defer conn.Close()

	// Listen on remote server port, process will hang at here
	listener, err := conn.Listen("tcp", remoteEndpoint)
	if err != nil {
		log.Error().Err(err).Msgf("Fail to listen remote endpoint")
		return err
	}
	defer listener.Close()

	log.Info().Msgf("Forward %s to localEndpoint %s", remoteEndpoint, localEndpoint)

	// handle incoming connections on reverse forwarded tunnel
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to accept remote request")
			return err
		}

		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		local, err := net.Dial("tcp", localEndpoint)
		if err != nil {
			_ = client.Close()
			log.Error().Err(err).Msgf("Local service error")
		} else {
			go handleClient(client, local)
		}
	}
}

func connection(username string, password string, address string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Error().Err(err).Msgf("Fail create ssh connection")
	}
	return conn, err
}

func handleClient(client net.Conn, remote net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Error().Err(err).Msgf("Error while copy remote->local")
		}
		wg.Done()
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Error().Err(err).Msgf("Error while copy local->remote")
		}
		wg.Done()
	}()

	wg.Wait()
	err := remote.Close()
	if err != nil {
		log.Error().Err(err).Msgf("Close connection failed")
	}
	err = client.Close()
	if err != nil {
		log.Error().Err(err).Msgf("Close connection failed")
	}
}
