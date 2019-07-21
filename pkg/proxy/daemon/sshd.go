package daemon

import (
	"os/exec"

	"github.com/rs/zerolog/log"
)

// StartSSHDaemon start sshd daemon
func StartSSHDaemon() (err error) {
	cmd := exec.Command("/usr/sbin/sshd", "-D")
	err = cmd.Start()
	if err != nil {
		return
	}
	pid := cmd.Process.Pid
	log.Info().Msgf("SSHD start at pid: %d\n", pid)
	go func() {
		err = cmd.Wait()
		log.Error().Msgf("SSHD Exited with error: %v\n", err)
	}()
	return
}
