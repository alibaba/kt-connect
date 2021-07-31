package exec

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshtunnelling"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
)

// CliInterface ...
type CliInterface interface {
	Kubectl() kubectl.CliInterface
	SSHUttle() sshuttle.CliInterface
	SSH() ssh.CliInterface
	SSHTunnelling() sshtunnelling.CliInterface
}

// Cli ...
type Cli struct {
	KubeOptions []string

	TunName  string
	SourceIP string
	DestIP   string
}

// Kubectl ...
func (c *Cli) Kubectl() kubectl.CliInterface {
	return &kubectl.Cli{KubeOptions: c.KubeOptions}
}

// SSHUttle ...
func (c *Cli) SSHUttle() sshuttle.CliInterface {
	return &sshuttle.Cli{}
}

// SSH ...
func (c *Cli) SSH() ssh.CliInterface {
	return &ssh.Cli{}
}

func (c *Cli) SSHTunnelling() sshtunnelling.CliInterface {
	return &sshtunnelling.Cli{
		TunName:  c.TunName,
		SourceIP: c.SourceIP,
		DestIP:   c.DestIP,
	}
}
