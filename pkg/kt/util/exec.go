package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// SSHRemotePortForward ssh remote port forward
func SSHRemotePortForward(localPort string, remoteHost string, remotePort string, remoteSSHPort int) *exec.Cmd {
	return exec.Command("ssh",
		"-oStrictHostKeyChecking=no",
		"-oUserKnownHostsFile=/dev/null",
		"-i", PrivateKeyPath(),
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
		"-i", PrivateKeyPath(),
		"-D", fmt.Sprintf("%d", proxyPort),
		fmt.Sprintf("root@%s", remoteHost), "-p"+fmt.Sprintf("%d", remoteSSHPort),
		"sh", "loop.sh",
	)
}

// PortForward kubectl port forward
func PortForward(kubeConifg string, namespace string, resource string, remotePort int) *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+kubeConifg, "-n", namespace, "port-forward", resource, fmt.Sprintf("%d", remotePort)+":22")
}

// SSHUttle ssh-baed vpn connect
func SSHUttle(remoteHost string, remotePort int, DNSServer string, disableDNS bool, cidrs []string, debug bool) *exec.Cmd {
	args := []string{}
	if !disableDNS {
		args = append(args, "--dns", "--to-ns", DNSServer)
	}

	if debug {
		args = append(args, "-v")
	}

	subCommand := fmt.Sprintf("ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -i %s", PrivateKeyPath())
	args = append(args, "-e", subCommand, "-r", fmt.Sprintf("root@%s:%d", remoteHost, remotePort), "-x", remoteHost)
	args = append(args, cidrs...)
	return exec.Command("sshuttle", args...)
}

// BackgroundRun run cmd in background
func BackgroundRun(cmd *exec.Cmd, name string, debug bool) (err error) {

	log.Debug().Msgf("Child, os.Args = %+v", os.Args)
	log.Debug().Msgf("Child, cmd.Args = %+v", cmd.Args)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()
	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Start()

	if err != nil {
		return
	}

	go func() {
		err = cmd.Wait()
		log.Printf("%s exited", name)
	}()

	time.Sleep(time.Duration(2) * time.Second)
	pid := cmd.Process.Pid
	log.Printf("%s start at pid: %d", name, pid)
	return
}
