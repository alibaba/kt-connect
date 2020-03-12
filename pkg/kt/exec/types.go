package exec

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
)

// CliInterface ...
type CliInterface interface {
	KubectlInterface() kubectl.CliInterface
	SSHUttleInterface() sshuttle.CliInterface
	SSHInterface() ssh.CliInterface
}

// Cli ...
type Cli struct {
	KubeConfig string
}

// KubectlInterface ...
func (c *Cli) KubectlInterface() kubectl.CliInterface {
	return &kubectl.Cli{KubeConfig: c.KubeConfig}
}

// SSHUttleInterface ...
func (c *Cli) SSHUttleInterface() sshuttle.CliInterface {
	return &sshuttle.Cli{}
}

// SSHInterface ...
func (c *Cli) SSHInterface() ssh.CliInterface {
	return &ssh.Cli{}
}
