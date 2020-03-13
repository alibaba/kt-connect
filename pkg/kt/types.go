package kt

import (
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
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
	return &exec.Cli{
		KubeConfig: c.Options.KubeConfig,
	}
}
