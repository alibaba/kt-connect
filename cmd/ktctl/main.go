package main

import (
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"os"
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
	app.Authors = general.NewCliAuthor()
	app.Flags = general.AppFlags(options, version)

	context := &kt.Cli{Options: options}
	action := command.Action{}

	app.Commands = newCommands(context, &action, options)
	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msgf("End with error: %s", err.Error())
		general.CleanupWorkspace(context, options)
		os.Exit(-1)
	}
}

// NewCommands return new Connect Action
func newCommands(kt kt.CliInterface, action command.ActionInterface, options *opt.DaemonOptions) []cli.Command {
	return []cli.Command{
		command.NewConnectCommand(kt, options, action),
		command.NewExchangeCommand(kt, options, action),
		command.NewMeshCommand(kt, options, action),
		command.NewProvideCommand(kt, options, action),
		command.NewCleanCommand(kt, options, action),
		command.NewDashboardCommand(kt, options, action),
	}
}
