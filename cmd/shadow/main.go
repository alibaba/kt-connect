package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/proxy/dnsserver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	log.Info().Msg("shadow staring...")
	srv := dnsserver.NewDNSServerDefault()
	err := srv.ListenAndServe()
	if err != nil {
		log.Error().Msg(err.Error())
		panic(err.Error())
	}
	log.Info().Msg("shadow(DNS) start at 53 successful")
	log.Info().Msg("shadow start successful")
}
