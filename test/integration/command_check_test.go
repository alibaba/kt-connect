package integration

import (
	"testing"
	"github.com/urfave/cli"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/command"
)

func Test_should(t *testing.T) {
	args := []string{"ktctl", "check"}
	options := options.NewDaemonOptions()
	app := cli.NewApp()
	app.Commands = []cli.Command {
		command.NewCheckCommand(options),
	} 
	err := app.Run(args)
	if err != nil {
		t.Errorf("%s failed: %v", args, err)
	}
}