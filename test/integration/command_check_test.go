package integration

import (
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/urfave/cli"
)

func Test_shouldCheckLocalDenpendencies(t *testing.T) {
	args := []string{"ktctl", "check"}
	options := options.NewDaemonOptions()
	app := cli.NewApp()
	app.Commands = []cli.Command{
		command.NewCheckCommand(options),
	}
	err := app.Run(args)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}
}
