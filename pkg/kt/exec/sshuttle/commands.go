package sshuttle

import (
	"fmt"
	"os/exec"
)

// Version check sshuttle version
func Version() *exec.Cmd {
	return exec.Command("sshuttle", "--version")
}

// SSHUttle ssh-baed vpn connect
func SSHUttle(remoteHost, privateKeyPath string, remotePort int, DNSServer string, disableDNS bool, cidrs []string, debug bool) *exec.Cmd {
	args := []string{}
	if !disableDNS {
		args = append(args, "--dns", "--to-ns", DNSServer)
	}

	if debug {
		args = append(args, "-v")
	}

	subCommand := fmt.Sprintf("ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i %s", privateKeyPath)
	args = append(args, "-e", subCommand, "-r", fmt.Sprintf("root@%s:%d", remoteHost, remotePort), "-x", remoteHost)
	args = append(args, cidrs...)
	return exec.Command("sshuttle", args...)
}
