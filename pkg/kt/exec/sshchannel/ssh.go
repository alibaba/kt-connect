package sshchannel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wzshiming/socks5"
	"golang.org/x/crypto/ssh"
)

// SSHChannel ssh channel
type SSHChannel struct{}

type SocksLogger struct {}

func (s SocksLogger) Println(v ...interface{}) {
	log.Info().Msgf(fmt.Sprint(v...))
}

// StartSocks5Proxy start socks5 proxy
func (c *SSHChannel) StartSocks5Proxy(privateKey string, sshAddress, socks5Address string) (err error) {
	conn, err := connection(privateKey, sshAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	svc := &socks5.Server{
		Logger: SocksLogger{},
		ProxyDial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			return conn.Dial(network, address)
		},
	}
	return svc.ListenAndServe("tcp", socks5Address)
}

// RunScript run the script on remote host.
func (c *SSHChannel) RunScript(privateKey string, sshAddress, script string) (result string, err error) {
	conn, err := connection(privateKey, sshAddress)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create ssh tunnel")
		return "", err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create ssh session")
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
func (c *SSHChannel) ForwardRemoteToLocal(privateKey string, sshAddress, remoteEndpoint, localEndpoint string) (err error) {
	conn, err := connection(privateKey, sshAddress)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create ssh tunnel")
		return err
	}

	// Listen on remote server port of shadow pod, via ssh connection
	listener, err := conn.Listen("tcp", remoteEndpoint)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to listen remote endpoint")
		return err
	}

	log.Info().Msgf("Forward %s to local endpoint %s", remoteEndpoint, localEndpoint)

	// Handle incoming connections on reverse forwarded tunnel
	go handleRequest(listener, localEndpoint)
	return
}

func handleRequest(listener net.Listener, localEndpoint string) {
	for {
		// Wait requests from remote endpoint
		client, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msgf("Failed to accept remote request")
		}

		// Open a (local) connection to localEndpoint whose content will be forwarded to remoteEndpoint
		local, err := net.Dial("tcp", localEndpoint)
		if err != nil {
			_ = client.Close()
			log.Error().Err(err).Msgf("Local service error")
		} else {
			// Handle request in individual coroutine, current coroutine continue to accept more requests
			go handleClient(client, local)
		}
	}
}

func connection(privateKey string, address string) (*ssh.Client, error) {
	key, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
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
