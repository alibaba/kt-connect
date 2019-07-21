package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/proxy/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	err := server.Run()
	if err != nil {
		panic(err.Error())
	}
}
