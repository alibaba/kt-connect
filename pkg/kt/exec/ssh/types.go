package ssh

import "os/exec"

// CliInterface ...
type CliInterface interface {
	Version() *exec.Cmd
	TunnelToRemote(localTun int, remoteHost, privateKeyPath string, remoteSSHPort int) *exec.Cmd
}

// Cli ...
type Cli struct{}
