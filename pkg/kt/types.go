package kt

import (
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// CliInterface ...
type CliInterface interface {
	KubernetesInterface() cluster.KubernetesInterface
	ShadowInterface() connect.ShadowInterface
	ExecInterface() exec.CliInterface
}

// Cli ...
type Cli struct {
	Options *options.DaemonOptions
}

// KubernetesInterface ...
func (c *Cli) KubernetesInterface() cluster.KubernetesInterface {
	return &cluster.Kubernetes{
		KubeConfig: c.Options.KubeConfig,
	}
}

// ShadowInterface ...
func (c *Cli) ShadowInterface() connect.ShadowInterface {
	return &connect.Shadow{
		Options: c.Options,
	}
}

// ExecInterface ...
func (c *Cli) ExecInterface() exec.CliInterface {
	return &exec.Cli{
		KubeConfig: c.Options.KubeConfig,
	}
}
