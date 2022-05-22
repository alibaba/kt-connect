package sshuttle

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"io"
	"os"
	"os/exec"
	"strings"
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
func (s *Cli) Connect(req *SSHVPNRequest) *exec.Cmd {
	var args []string
	if opt.Get().Connect.DnsMode == util.DnsModePodDns {
		args = append(args, "--dns", "--to-ns", req.RemoteDNSServerAddress)
	}
	if opt.Get().Global.Debug {
		args = append(args, "--verbose")
	}

	subCommand := fmt.Sprintf("ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i %s", req.RemoteSSHPKPath)
	remoteAddr := fmt.Sprintf("root@%s:%d", common.Localhost, req.LocalSshPort)
	args = append(args, "--ssh-cmd", subCommand, "--remote", remoteAddr, "--exclude", common.Localhost)
	if opt.Get().Connect.ExcludeIps != "" {
		for _, ip := range strings.Split(opt.Get().Connect.ExcludeIps, ",") {
			args = append(args, "--exclude", ip)
		}
	}
	args = append(args, req.CustomCIDR...)
	cmd := exec.Command("sshuttle", args...)
	if !opt.Get().Global.Debug {
		stdoutPipe, _ := cmd.StdoutPipe()
		stderrPipe, _ := cmd.StderrPipe()
		if stdoutPipe != nil && stderrPipe != nil {
			go io.Copy(io.Discard, stdoutPipe)
			go io.Copy(io.Discard, stderrPipe)
		}
	}
	return cmd
}
