package main

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/proxy/dnsserver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if level, err := zerolog.ParseLevel(os.Getenv(common.EnvVarLogLevel)); err != nil {
		log.Error().Err(err).Msgf("Failed to parse log level")
	} else {
		zerolog.SetGlobalLevel(level)
	}
}

func main() {
	log.Info().Msg("Shadow staring...")
	dnsserver.Start()
}
