package ssh

import "os/exec"

// CliInterface ...
type CliInterface interface {
	Version() *exec.Cmd
	ForwardRemoteRequestToLocal(localPort, remoteHost, remotePort, privateKeyPath string, remoteSSHPort int) *exec.Cmd
	DynamicForwardLocalRequestToRemote(remoteHost, privateKeyPath string, remoteSSHPort int, proxyPort int) *exec.Cmd
	TunnelToRemote(localTun int, remoteHost, privateKeyPath string, remoteSSHPort int) *exec.Cmd
}

// Cli ...
type Cli struct{}
