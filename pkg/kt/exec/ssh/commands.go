package ssh

import (
	"fmt"
	"os/exec"
)

// Version check sshuttle version
func (s *Cli) Version() *exec.Cmd {
	return exec.Command("ssh", "-V")
}

// ForwardRemoteRequestToLocal ssh remote port forward
func (s *Cli) ForwardRemoteRequestToLocal(localPort, remoteHost, remotePort, privateKeyPath string, remoteSSHPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", privateKeyPath,
		"-R", remotePort+":127.0.0.1:"+localPort,
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}

// DynamicForwardLocalRequestToRemote ssh remote port forward
func (s *Cli) DynamicForwardLocalRequestToRemote(remoteHost, privateKeyPath string, remoteSSHPort int, proxyPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", privateKeyPath,
		"-D", fmt.Sprintf("%d", proxyPort),
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}
