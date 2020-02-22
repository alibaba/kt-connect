package ssh

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"os/exec"
)

// SSHVersion check sshuttle version
func Version() *exec.Cmd {
	return exec.Command("ssh", "-V")
}

// SSHRemotePortForward ssh remote port forward
func SSHRemotePortForward(localPort string, remoteHost string, remotePort string, remoteSSHPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", util.PrivateKeyPath(),
		"-R", remotePort+":127.0.0.1:"+localPort,
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}

// SSHDynamicPortForward ssh remote port forward
func SSHDynamicPortForward(remoteHost string, remoteSSHPort int, proxyPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", util.PrivateKeyPath(),
		"-D", fmt.Sprintf("%d", proxyPort),
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}
