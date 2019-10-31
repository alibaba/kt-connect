package util

import (
	"fmt"
	"os"
	"os/exec"
)

// IsDaemonRunning check daemon is running or not
func IsDaemonRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// HomeDir Current User home dir
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	return "/root"
}

// PrivateKeyPath Get ssh private key path
func PrivateKeyPath() string {
	userHome := HomeDir()
	privateKey := fmt.Sprintf("%s/.kt_id_rsa", userHome)
	return privateKey
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

// GetRandomSSHPort get pod random ssh port
func GetRandomSSHPort(podIP string) string {
	return fmt.Sprintf("22%s", podIP[len(podIP)-2:len(podIP)])
}

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
	args = append(args, "-v", "-e", subCommand, "-r", fmt.Sprintf("root@%s:%d", remoteHost, remotePort), "-x", remoteHost)
	args = append(args, cidrs...)
	return exec.Command("sshuttle", args...)
}
