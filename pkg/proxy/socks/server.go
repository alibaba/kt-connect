package socks

import (
	"fmt"
	"log"
	"os"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/wzshiming/socks4"
)

func Start() {
	logger := log.New(os.Stderr, "[socks4] ", log.LstdFlags)
	svc := &socks4.Server{
		Logger: logger,
	}
	err := svc.ListenAndServe("tcp", fmt.Sprintf(":%d", common.Socks4Port))
	if err != nil {
		logger.Println(err)
	}
}
