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
	clientset, err := cluster.GetKubernetesClient(c.Options.KubeConfig)
	if err != nil {
		return nil, err
	}
	return &cluster.Kubernetes{
		KubeConfig: c.Options.KubeConfig,
		Clientset:  clientset,
	}, nil
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
