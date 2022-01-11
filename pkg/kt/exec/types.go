package exec

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec/dns"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/exec/tun"
)

// CliInterface ...
type CliInterface interface {
	Sshuttle() sshuttle.Sshuttle
	Tunnel() tun.Tunnel
	SshChannel() sshchannel.Channel
	DnsConfig() dns.DnsConfig
}

// Cli ...
type Cli struct {}

// SshChannel ...
func (c *Cli) SshChannel() sshchannel.Channel {
	return &sshchannel.SSHChannel{}
}

// Sshuttle ...
func (c *Cli) Sshuttle() sshuttle.Sshuttle {
	return &sshuttle.Cli{}
}

// Tunnel commands for tun device
func (c *Cli) Tunnel() tun.Tunnel {
	return &tun.Cli{}
}

// DnsConfig commands for dns configure
func (c *Cli) DnsConfig() dns.DnsConfig {
	return &dns.Cli{}
}
