package socks

import (
	"log"
	"os"

	"github.com/wzshiming/socks4"
)

func Start() {
	logger := log.New(os.Stderr, "[socks4] ", log.LstdFlags)
	svc := &socks4.Server{
		Logger: logger,
	}
	svc.Authentication = socks4.UserAuth("kt")
	err := svc.ListenAndServe("tcp", ":1080")
	if err != nil {
		logger.Println(err)
	}
}
