package channel

import (
	"log"
	"testing"
)

func Test_exec(t *testing.T) {
	conn, err := connection()
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	exec(conn, "ls -al")
}

func Test_dynamicPortForward(t *testing.T) {
	conn, err := connection()
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	dynamicPortForward(conn)
}

func Test_forwardRemoteToLocal(t *testing.T) {
	conn, err := connection()
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	forwardRemoteToLocal(conn)
}
