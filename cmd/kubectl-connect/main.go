package main

import (
	"os"

	"github.com/alibaba/kt-connect/pkg/kt"

	"github.com/alibaba/kt-connect/pkg/kt/command"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	options := opt.NewDaemonOptions()
	context := &kt.Cli{Options: options}

	app := cli.NewApp()
	app.Name = "connect"
	app.Usage = "connect to cluster with vpn or socks5 proxy"
	app.Version = "0.0.1"
	app.Authors = command.NewCliAuthor()
	app.Flags = command.PluginConnectFlags(options)
	app.Action = func(c *cli.Context) error {
		action := command.Action{}
		if options.Debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		return action.Connect(context, options)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error().Msg(err.Error())
		command.CleanupWorkspace(context, options)
		os.Exit(-1)
	}
}
