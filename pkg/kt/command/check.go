package command

import (

	osexec "os/exec"

	"runtime"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"

	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
)

// NewCheckCommand return new check command
func NewCheckCommand(options *options.DaemonOptions) cli.Command {
	return cli.Command{
		Name:  "check",
		Usage: "check local dependency for ktctl",
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			return Check(options)
		},
	}
}

// Check check local denpendency for kt connect
func Check(options *options.DaemonOptions) (err error) {
	log.Info().Msgf("system info %s-%s", runtime.GOOS, runtime.GOARCH)

	err = runCommandWithMsg(
		ssh.Version(),
		"checking ssh version", "ssh is missing, please make sure command ssh is work right at your local first", 
	)

	if err != nil {
		return
	}

	err =  runCommandWithMsg(
		kubectl.Version(options.KubeConfig),
		"checking kubectl version", "kubectl is missing, please make sure kubectl is working right at your local first", 
	)

	if err != nil {
		return
	}

	err = runCommandWithMsg(
		sshuttle.Version(),
		"checking sshuttle version", "sshuttle is missing, you can only use 'ktctl connect --method socks5' with Socks5 proxy mode",
	)

	if err != nil {
		return
	}

	log.Info().Msg("KT Connect is ready, enjoy it!")
	return nil
}

func runCommandWithMsg(cmd *osexec.Cmd, title string, msg string) (err error) {
	log.Info().Msg(title)
	err = exec.RunAndWait(cmd, title, true)
	if err != nil {
		log.Warn().Msg(msg)
	}
	return
}