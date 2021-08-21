package sshuttle

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

// Version check sshuttle version
func (s *Cli) Version() *exec.Cmd {
	return exec.Command("sshuttle", "--version")
}

// Connect ssh-baed vpn connect
func (s *Cli) Connect(remoteHost, privateKeyPath string, remotePort int, DNSServer string, disableDNS bool, cidrs []string, debug bool) *exec.Cmd {
	var args []string
	if !disableDNS {
		args = append(args, "--dns", "--to-ns", DNSServer)
	}

	if debug {
		args = append(args, "--verbose")
	}

	subCommand := fmt.Sprintf("ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i %s", privateKeyPath)
	remoteAddr := fmt.Sprintf("root@%s:%d", remoteHost, remotePort)
	args = append(args, "--ssh-cmd", subCommand, "--remote", remoteAddr, "--exclude", remoteHost)
	args = append(args, cidrs...)
	cmd := exec.Command("sshuttle", args...)
	if !debug {
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()
		if stdoutPipe != nil && stderrPipe != nil {
			go io.Copy(bufio.NewWriter(nil), stdoutPipe)
			go io.Copy(bufio.NewWriter(nil), stderrPipe)
		}
	}
	return cmd
}
