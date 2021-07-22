package socks

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/linfan/socks4"
	"github.com/rs/zerolog/log"
)

func Start() {
	svc := &socks4.Server{}
	err := svc.ListenAndServe("tcp", fmt.Sprintf(":%d", common.Socks4Port))
	if err != nil {
		log.Error().Err(err)
	}
}
