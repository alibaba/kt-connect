package exec

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/exec/tun"
)

// CliInterface ...
type CliInterface interface {
	Sshuttle() sshuttle.CliInterface
	Tunnel() tun.CliInterface
	SshChannel() sshchannel.Channel
}

// Cli ...
type Cli struct {}

// SshChannel ...
func (c *Cli) SshChannel() sshchannel.Channel {
	return &sshchannel.SSHChannel{}
}

// Sshuttle ...
func (c *Cli) Sshuttle() sshuttle.CliInterface {
	return &sshuttle.Cli{}
}

func (c *Cli) Tunnel() tun.CliInterface {
	return &tun.Cli{}
}
