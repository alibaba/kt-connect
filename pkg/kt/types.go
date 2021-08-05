package kt

import (
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"strings"
)

// CliInterface ...
type CliInterface interface {
	Kubernetes() (cluster.KubernetesInterface, error)
	Shadow() connect.ShadowInterface
	Exec() exec.CliInterface
}

// Cli ...
type Cli struct {
	Options *options.DaemonOptions
}

// Kubernetes ...
func (c *Cli) Kubernetes() (cluster.KubernetesInterface, error) {
	if c.Options.RuntimeOptions.Clientset != nil {
		return cluster.CreateFromClientSet(c.Options.RuntimeOptions.Clientset)
	}
	return cluster.Create(c.Options.KubeConfig)
}

// Shadow ...
func (c *Cli) Shadow() connect.ShadowInterface {
	return &connect.Shadow{
		Options: c.Options,
	}
}

// Exec ...
func (c *Cli) Exec() exec.CliInterface {
	maskLen := c.Options.ConnectOptions.TunCidr[strings.Index(c.Options.ConnectOptions.TunCidr, "/")+1:]
	return &exec.Cli{
		KubeOptions: c.Options.KubeOptions,
		TunName:     c.Options.ConnectOptions.TunName,
		SourceIP:    c.Options.ConnectOptions.SourceIP,
		DestIP:      c.Options.ConnectOptions.DestIP,
		MaskLen:     maskLen,
	}
}
