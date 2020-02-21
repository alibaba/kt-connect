package util

import (
	"fmt"
	"os/exec"
)

// ApplyDashboardToCluster Apply Dashboard to cluster
func ApplyDashboardToCluster() *exec.Cmd {
	return exec.Command(
		"kubectl",
		"-n",
		"kube-system",
		"apply",
		"-f",
		"https://raw.githubusercontent.com/alibaba/kt-connect/master/docs/deploy/manifest/all-in-one.yaml")
}

//PortForwardDashboardToLocal forward dashboardto local
func PortForwardDashboardToLocal(port string) *exec.Cmd {
	return exec.Command(
		"kubectl",
		"-n",
		"kube-system",
		"port-forward",
		"service/kt-dashboard",
		port+":80",
	)
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

// KubectlVersion kubectl version
func KubectlVersion(kubeConifg string) *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+kubeConifg, "version", "--short", "port-forward")
}

// SSHUttleVersion check sshuttle version
func SSHUttleVersion() *exec.Cmd {
	return exec.Command("sshutle", "--version")
}

// SSHVersion check sshuttle version
func SSHVersion() *exec.Cmd {
	return exec.Command("ssh", "-V")
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
