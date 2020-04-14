package kubectl

import (
	"fmt"
	"os/exec"
)

// KUBECTL the path to kubectl
var KUBECTL = "kubectl"

// Version ...
func (k *Cli) Version() *exec.Cmd {
	args := kubectl(k, "")
	args = append(args, "version", "--short", "port-forward")
	return exec.Command(KUBECTL,
		args...,
	)
}

// ApplyDashboardToCluster ...
func (k *Cli) ApplyDashboardToCluster() *exec.Cmd {
	args := kubectl(k, "kube-system")
	args = append(
		args,
		"apply",
		"-f",
		"https://raw.githubusercontent.com/alibaba/kt-connect/master/docs/deploy/manifest/all-in-one.yaml")
	return exec.Command(
		KUBECTL,
		args...)
}

// PortForwardDashboardToLocal ...
func (k *Cli) PortForwardDashboardToLocal(port string) *exec.Cmd {
	args := kubectl(k, "kube-system")
	args = append(args, "port-forward",
		"service/kt-dashboard",
		port+":80")
	return exec.Command(
		KUBECTL,
		args...,
	)
}

// PortForward ...
func (k *Cli) PortForward(namespace, resource string, remotePort int) *exec.Cmd {
	args := kubectl(k, namespace)
	args = append(args, "port-forward",
		resource,
		fmt.Sprintf("%d", remotePort)+":22")
	return exec.Command(
		KUBECTL,
		args...,
	)
}

func kubectl(k *Cli, namespace string) []string {
	var args []string

	if k.KubeConfig != "" {
		args = append(args, "--kubeconfig="+k.KubeConfig)
	}

	if namespace != "" {
		args = append(args, "-n")
		args = append(args, namespace)
	}

	return args
}
