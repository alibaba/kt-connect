package kubectl

import (
	"fmt"
	"os/exec"
)

// KubectlVersion kubectl version
func KubectlVersion(kubeConifg string) *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+kubeConifg, "version", "--short", "port-forward")
}

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

// PortForward kubectl port forward
func PortForward(kubeConifg string, namespace string, resource string, remotePort int) *exec.Cmd {
	return exec.Command("kubectl", "--kubeconfig="+kubeConifg, "-n", namespace, "port-forward", resource, fmt.Sprintf("%d", remotePort)+":22")
}
