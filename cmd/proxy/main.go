package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/proxy/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := server.Run()
	if err != nil {
		panic(err.Error())
	}
}
