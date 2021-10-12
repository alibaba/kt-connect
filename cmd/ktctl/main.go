package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/kt/util"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	version = "dev"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: util.IsWindows()})
}

func main() {
	options := opt.NewDaemonOptions(version)

	app := cli.NewApp()
	app.Name = "KtConnect"
	app.Usage = ""
	app.Version = version
	app.Authors = command.NewCliAuthor()
	app.Flags = command.AppFlags(options, version)

	context := &kt.Cli{Options: options}
	action := command.Action{}

	app.Commands = command.NewCommands(context, &action, options)
	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msgf("End with error: %s", err.Error())
		command.CleanupWorkspace(context, options)
		os.Exit(-1)
	}
}
