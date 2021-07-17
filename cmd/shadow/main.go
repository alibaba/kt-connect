package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/proxy/dnsserver"
	"github.com/alibaba/kt-connect/pkg/proxy/socks"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	connectMethod := os.Getenv(common.EnvVarConnectMethod)
	log.Info().Msg("shadow staring...")
	if connectMethod == common.ConnectMethodSocks {
		socks.Start()
	} else {
		dnsserver.Start()
	}
}
