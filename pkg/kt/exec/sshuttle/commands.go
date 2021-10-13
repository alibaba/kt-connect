package sshuttle

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"io"
	"os"
	"os/exec"
)

// Version check sshuttle version
func (s *Cli) Version() *exec.Cmd {
	return exec.Command("sshuttle", "--version")
}

// Install try to install sshuttle
func (s *Cli) Install() *exec.Cmd {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		return exec.Command("sudo", "-u", sudoUser, "pip3", "install", "sshuttle")
	} else {
		return exec.Command("pip3", "install", "sshuttle")
	}
}

// Connect ssh-based vpn connect
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
	for _, ip := range util.GetLocalIps() {
		args = append(args, "--exclude", ip)
	}
	args = append(args, cidrs...)
	cmd := exec.Command("sshuttle", args...)
	if !debug {
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()
		if stdoutPipe != nil && stderrPipe != nil {
			go io.Copy(io.Discard, stdoutPipe)
			go io.Copy(io.Discard, stderrPipe)
		}
	}
	return cmd
}
