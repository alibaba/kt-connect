package kubectl

import "os/exec"

// CliInterface ...
type CliInterface interface {
	PortForward(namespace, resource string, remotePort, localPort int) *exec.Cmd
}

// Cli ...
type Cli struct {
	KubeOptions []string
}
