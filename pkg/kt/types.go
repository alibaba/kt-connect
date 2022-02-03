package kt

import (
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
)

// CliInterface ...
type CliInterface interface {
	Kubernetes() cluster.KubernetesInterface
}

// Cli ...
type Cli struct {}

// Kubernetes ...
func (c *Cli) Kubernetes() cluster.KubernetesInterface {
	return &cluster.Kubernetes{
		Clientset: opt.Get().RuntimeOptions.Clientset,
	}
}
