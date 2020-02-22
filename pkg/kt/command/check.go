package command

import (
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

// Check check local denpendency for kt connect
func (action *Action) Check(options *options.DaemonOptions) error {
	log.Info().Msgf("system info %s-%s", runtime.GOOS, runtime.GOARCH)

	log.Info().Msg("checking ssh version")
	command := ssh.SSHVersion()
	err := exec.RunAndWait(command, "ssh version", true)
	if err != nil {
		log.Error().Msg("ssh is missing, please make sure command ssh is work right at your local first")
		return err
	}

	log.Info().Msg("checking kubectl version")
	command = kubectl.KubectlVersion(options.KubeConfig)
	err = exec.RunAndWait(command, "kubectl version", true)
	if err != nil {
		log.Error().Msg("kubectl is missing, please make sure kubectl is working right at your local first")
		return err
	}

	log.Info().Msg("checking sshuttle version")
	command = sshuttle.SSHUttleVersion()
	err1 := exec.RunAndWait(command, "sshuttle version", true)
	if err1 != nil {
		log.Warn().Msg("sshuttle is missing, you can only use 'ktctl connect --method socks5' with Socks5 proxy mode")
	}

	log.Info().Msg("KT Connect is ready, enjoy it!")
	return nil
}
