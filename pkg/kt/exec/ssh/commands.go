package ssh

import (
	"fmt"
	"os/exec"
)

// Version check sshuttle version
func (s *Cli) Version() *exec.Cmd {
	return exec.Command("ssh", "-V")
}

func (s *Cli) TunnelToRemote(localTun int, remoteHost, privateKeyPath string, remoteSSHPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", privateKeyPath,
		"-f",
		"-w",
		fmt.Sprintf("%d:1", localTun),
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"true",
	)
}
