package kubectl

import (
	"os/exec"
	"strings"
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

func kubectl(k *Cli, namespace string) []string {
	var (
		args             []string
		isNamespaceField bool
	)
	if len(k.KubeOptions) != 0 {
		for _, opt := range k.KubeOptions {
			// instead namespace options
			isNamespaceField = strings.Contains(opt, "-n") || strings.Contains(opt, "--namespace")
			if isNamespaceField && namespace != "" {
				args = append(args, "-n", namespace)
				continue
			}
			args = append(args, strings.Fields(opt)...)
		}
	}
	return args
}
