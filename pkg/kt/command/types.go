package command

import (
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// ActionInterface all action defined
type ActionInterface interface {
	Connect(cli kt.CliInterface, options *options.DaemonOptions) error
	Preview(serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error
	Exchange(resourceName string, cli kt.CliInterface, options *options.DaemonOptions) error
	Mesh(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error
	Clean(cli kt.CliInterface, options *options.DaemonOptions) error
}

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}
