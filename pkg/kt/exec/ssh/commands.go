package ssh

import (
	"fmt"
	"os/exec"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// Version check sshuttle version
func Version() *exec.Cmd {
	return exec.Command("ssh", "-V")
}

// ForwardRemoteRequestToLocal ssh remote port forward
func ForwardRemoteRequestToLocal(localPort, remoteHost, remotePort string, remoteSSHPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", util.PrivateKeyPath(),
		"-R", remotePort+":127.0.0.1:"+localPort,
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}

// DynamicForwardLocalRequestToRemote ssh remote port forward
func DynamicForwardLocalRequestToRemote(remoteHost string, remoteSSHPort int, proxyPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", util.PrivateKeyPath(),
		"-D", fmt.Sprintf("%d", proxyPort),
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}
