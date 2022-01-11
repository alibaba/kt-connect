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

// Cli ...
type Cli struct{}

// SSHVPNRequest ...
type SSHVPNRequest struct {
	RemoteSSHHost          string
	RemoteSSHPKPath        string
	RemoteDNSServerAddress string
	CustomCIDR             []string
	Stop                   chan struct{}
	Debug                  bool
}
