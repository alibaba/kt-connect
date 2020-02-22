package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/urfave/cli"
)

// NewCheckCommand return new check command
func newCheckCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "check",
		Usage: "check local dependency for ktctl",
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			action := Action{}
			return action.Check(options)
		},
	}
}
