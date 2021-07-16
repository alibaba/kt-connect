package shadowsocks

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"log"
	"os"

	"github.com/wzshiming/shadowsocks"
)

func Start() {
	logger := log.New(os.Stderr, "[shadowsocks] ", log.LstdFlags)
	go func() {
		svc := &shadowsocks.Server{
			Logger:   logger,
			Cipher:   "chacha20-ietf-poly1305",
			Password: "kt-connect",
		}

		err := svc.ListenAndServe("tcp", fmt.Sprintf(":%d", common.ShadowSocksPort))
		if err != nil {
			logger.Println(err)
		}
		os.Exit(1)
	}()
	<-make(chan struct{})
}
