package kubectl

import (
	"fmt"
	"os/exec"
)

// Version ...
func (k *Cli) Version() *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+k.KubeConfig, "version", "--short", "port-forward")
}

// ApplyDashboardToCluster ...
func (k *Cli) ApplyDashboardToCluster() *exec.Cmd {
	return exec.Command(
		"kubectl",
		"--kubeconfig="+k.KubeConfig,
		"-n",
		"kube-system",
		"apply",
		"-f",
		"https://raw.githubusercontent.com/alibaba/kt-connect/master/docs/deploy/manifest/all-in-one.yaml")
}

// PortForwardDashboardToLocal ...
func (k *Cli) PortForwardDashboardToLocal(port string) *exec.Cmd {
	return exec.Command(
		"kubectl",
		"--kubeconfig="+k.KubeConfig,
		"-n",
		"kube-system",
		"port-forward",
		"service/kt-dashboard",
		port+":80",
	)
}

// PortForward ...
func (k *Cli) PortForward(namespace, resource string, remotePort int) *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+k.KubeConfig, "-n", namespace, "port-forward", resource, fmt.Sprintf("%d", remotePort)+":22")
}
