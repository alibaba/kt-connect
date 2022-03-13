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
	"os/signal"
	"syscall"
)

var (
	version = "dev"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: util.IsWindows()})
}

func main() {
	// this line must go first
	opt.Get().RuntimeStore.Version = version
	ch := setupCloseHandler()

	app := cli.NewApp()
	app.Name = "KtConnect"
	app.Usage = ""
	app.Version = version
	app.Authors = general.NewCliAuthor()
	app.Flags = general.AppFlags(opt.Get())
	app.Commands = newCommands(ch)
	// must overwrite default error handler to perform graceful exit
	app.ExitErrHandler = func(context *cli.Context, err error) {
		if err != nil {
			log.Error().Err(err).Msgf("Failed to start")
		}
		general.CleanupWorkspace()
		os.Exit(1)
	}
	// process will hang here
	_ = app.Run(os.Args)
	general.CleanupWorkspace()
}

// NewCommands return new Connect Action
func newCommands(ch chan os.Signal) []cli.Command {
	action := &command.Action{}
	return []cli.Command{
		command.NewConnectCommand(action, ch),
		command.NewExchangeCommand(action, ch),
		command.NewMeshCommand(action, ch),
		command.NewPreviewCommand(action, ch),
		command.NewRecoverCommand(action),
		command.NewCleanCommand(action),
	}
}

// setupCloseHandler registry close handler
func setupCloseHandler() (ch chan os.Signal) {
	ch = make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	return ch
}
