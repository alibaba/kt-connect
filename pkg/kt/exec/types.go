package exec

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
)

// CliInterface ...
type CliInterface interface {
	Kubectl() kubectl.CliInterface
	SSHUttle() sshuttle.CliInterface
	SSH() ssh.CliInterface
}

// Cli ...
type Cli struct {
	KubeConfig string
}

// Kubectl ...
func (c *Cli) Kubectl() kubectl.CliInterface {
	return &kubectl.Cli{KubeConfig: c.KubeConfig}
}

// SSHUttle ...
func (c *Cli) SSHUttle() sshuttle.CliInterface {
	return &sshuttle.Cli{}
}

// SSH ...
func (c *Cli) SSH() ssh.CliInterface {
	return &ssh.Cli{}
}
