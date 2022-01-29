package kt

import (
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// CliInterface ...
type CliInterface interface {
	Kubernetes() cluster.KubernetesInterface
}

// Cli ...
type Cli struct {
	Options *options.DaemonOptions
}

// Kubernetes ...
func (c *Cli) Kubernetes() cluster.KubernetesInterface {
	return &cluster.Kubernetes{
		Clientset: c.Options.RuntimeOptions.Clientset,
	}
}
