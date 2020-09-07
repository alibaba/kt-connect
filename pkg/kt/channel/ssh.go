package channel

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/armon/go-socks5"
	"golang.org/x/net/context"

	"golang.org/x/crypto/ssh"
)

// DynamicPortForward create socks5 proxy
func DynamicPortForward(username string, password string, address string, socks5Address string) error {
	conn, err := connection(username, password, address)
	defer conn.Close()
	if err != nil {
		return err
	}

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(conf)

	if err != nil {
		return err
	}

	if err := serverSocks.ListenAndServe("tcp", socks5Address); err != nil {
		fmt.Println("failed to create socks5 server", err)
		return err
	}
	fmt.Println("dynamic port forward successful")
	return nil
}

// ForwardRemoteToLocal forward remote request to local
func ForwardRemoteToLocal(username string, password string, address string, remoteEndpoint string, localEndpoint string) {
	conn, err := connection(username, password, address)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Listen on remote server port
	listener, err := conn.Listen("tcp", remoteEndpoint)
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}
	defer listener.Close()

	// handle incoming connections on reverse forwarded tunnel
	for {
		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		local, err := net.Dial("tcp", localEndpoint)
		if err != nil {
			log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
		}

		client, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
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
		log.Fatal("fail create ssh connection", err)
	}
	return conn, err
}

func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println("error while copy remote->local:", err)
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(err)
		}
		chDone <- true
	}()

	<-chDone
}
