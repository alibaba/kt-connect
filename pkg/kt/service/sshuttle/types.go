package sshuttle

import (
	"os/exec"
)

// Sshuttle ...
type Sshuttle interface {
	Version() *exec.Cmd
	Install() *exec.Cmd
	Connect(req *SSHVPNRequest) *exec.Cmd
}

// SSHVPNRequest ...
type SSHVPNRequest struct {
	LocalSshPort           int
	RemoteSSHPKPath        string
	RemoteDNSServerAddress string
	CustomCIDR             []string
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() Sshuttle {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}
