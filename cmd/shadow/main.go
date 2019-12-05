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
	log.Info().Msg("Start kt connect proxy")
	srv := dnsserver.NewDNSServerDefault()
	err := srv.ListenAndServe()
	if err != nil {
		panic(err.Error())
	}
	log.Info().Msgf("DNS Server Start At 53...\n")
}
