package command

import (
	osexec "os/exec"

	"github.com/alibaba/kt-connect/pkg/kt"

	"runtime"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
)

// newCheckCommand return new check command
func newCheckCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "check",
		Usage: "check local dependency for ktctl",
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			return action.Check(cli)
		},
	}
}

// Check check local denpendency for KtConnect
func (action *Action) Check(cli kt.CliInterface) (err error) {
	log.Info().Msgf("System info %s-%s", runtime.GOOS, runtime.GOARCH)

	err = runCommandWithMsg(
		cli.Exec().SSH().Version(),
		"Checking ssh version", "Ssh is missing, please make sure command ssh is work right at your local first",
	)

	if err != nil {
		return
	}

	err = runCommandWithMsg(
		cli.Exec().Kubectl().Version(),
		"Checking kubectl version", "Kubectl is missing, please make sure kubectl is working right at your local first",
	)

	if err != nil {
		return
	}

	err = runCommandWithMsg(
		cli.Exec().Sshuttle().Version(),
		"Checking sshuttle version", "Sshuttle is missing, you can only use 'ktctl connect --method socks5' with Socks5 proxy mode",
	)

	if err != nil {
		return
	}

	log.Info().Msg("KtConnect is ready, enjoy it!")
	return nil
}

func runCommandWithMsg(cmd *osexec.Cmd, title string, msg string) (err error) {
	log.Info().Msg(title)
	err = exec.RunAndWait(cmd, title)
	if err != nil {
		log.Warn().Msg(msg)
	}
	return
}
