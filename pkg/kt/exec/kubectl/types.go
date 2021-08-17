package kubectl

import "os/exec"

// CliInterface ...
type CliInterface interface {
	Version() *exec.Cmd
	ApplyDashboardToCluster() *exec.Cmd
	PortForwardDashboardToLocal(port string) *exec.Cmd
	PortForward(namespace, resource string, remotePort, localPort int) *exec.Cmd
}

// Cli ...
type Cli struct {
	KubeOptions []string
}
