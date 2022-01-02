package kubectl

import (
	"fmt"
	"os/exec"
	"strings"
)

// KUBECTL the path to kubectl
var KUBECTL = "kubectl"

// PortForward ...
func (k *Cli) PortForward(namespace, resource string, remotePort, localPort int) *exec.Cmd {
	args := kubectl(k, namespace)
	args = append(args, "port-forward",
		resource,
		"--address=127.0.0.1",
		fmt.Sprintf("%d:%d", localPort, remotePort))
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
