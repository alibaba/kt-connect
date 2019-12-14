package main

import (
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"os"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	options := options.NewDaemonOptions()

	app := cli.NewApp()
	app.Name = "KT Connect"
	app.Usage = ""
	app.Version = "0.0.8"
	app.Authors = command.NewCliAuthor()
	app.Flags = command.AppFlags(options)
	app.Commands = command.NewCommands(options)
	command.SetUpCloseHandler(options)

	err := app.Run(os.Args)
	if err != nil {
		command.CleanupWorkspace(options)
		log.Fatal().Err(err)
	}
}
