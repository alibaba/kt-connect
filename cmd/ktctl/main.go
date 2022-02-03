package main

import (
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
	app := cli.NewApp()
	app.Name = "KtConnect"
	app.Usage = ""
	app.Version = version
	app.Authors = general.NewCliAuthor()
	app.Flags = general.AppFlags(opt.Get(), version)

	action := command.Action{}

	app.Commands = newCommands(&action)
	app.ExitErrHandler = func(context *cli.Context, err error) {
		log.Error().Err(err).Msgf("Failed to start")
	}
	if err := app.Run(os.Args); err != nil {
		general.CleanupWorkspace()
		os.Exit(-1)
	}
}

// NewCommands return new Connect Action
func newCommands(action command.ActionInterface) []cli.Command {
	return []cli.Command{
		command.NewConnectCommand(action),
		command.NewExchangeCommand(action),
		command.NewMeshCommand(action),
		command.NewPreviewCommand(action),
		command.NewCleanCommand(action),
	}
}
