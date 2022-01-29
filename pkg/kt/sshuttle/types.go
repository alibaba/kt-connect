package sshuttle

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"os/exec"
)

// Sshuttle ...
type Sshuttle interface {
	Version() *exec.Cmd
	Install() *exec.Cmd
	Connect(opt *options.ConnectOptions, req *SSHVPNRequest) *exec.Cmd
}

// SSHVPNRequest ...
type SSHVPNRequest struct {
	RemoteSSHHost          string
	RemoteSSHPKPath        string
	RemoteDNSServerAddress string
	CustomCIDR             []string
	Stop                   chan struct{}
	Debug                  bool
}

// Cli the singleton type
type Cli struct {}
var instance *Cli

// Ins get singleton instance
func Ins() *Cli {
	if instance == nil {
		instance = &Cli{}
	}
	return instance
}
