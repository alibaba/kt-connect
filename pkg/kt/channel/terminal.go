package channel

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
)

var (
	sshAddr        = "127.0.0.1:2222"
	localAddr      = "127.0.0.1:5000"
	remoteEndpoint = "0.0.0.0:8001"
	localEndpoint  = "127.0.0.1:8001"
	socks5Address  = "127.0.0.1:2223"
)

func connection() (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.Auth = []ssh.AuthMethod{ssh.Password("root")}

	conn, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		log.Fatal("创建ssh client 失败", err)
	}
	return conn, err
}

func dynamicPortForward(conn *ssh.Client) {
	for {
		conf := &socks5.Config{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return conn.Dial(network, addr)
			},
		}

		serverSocks, err := socks5.New(conf)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := serverSocks.ListenAndServe("tcp", socks5Address); err != nil {
			fmt.Println("failed to create socks5 server", err)
		}

		fmt.Println("dynamic port forward successful")
	}
}

func forwardRemoteToLocal(conn *ssh.Client) {
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

func exec(conn *ssh.Client, cmd string) {
	//创建ssh-session
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal("创建ssh session 失败", err)
	}
	defer session.Close()
	//执行远程命令

	combo, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Fatal("远程执行cmd 失败", err)
	}
	log.Println("命令输出:", string(combo))
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
