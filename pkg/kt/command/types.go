package command

import (
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// ActionInterface all action defined
type ActionInterface interface {
	OpenDashboard(cli kt.CliInterface, options *options.DaemonOptions) error
	Connect(cli kt.CliInterface, options *options.DaemonOptions) error
	Check(cli kt.CliInterface) error
	Run(service string, cli kt.CliInterface, options *options.DaemonOptions) error
	Exchange(service string, cli kt.CliInterface, options *options.DaemonOptions) error
	Mesh(service string, cli kt.CliInterface, options *options.DaemonOptions) error
	Clean(cli kt.CliInterface, options *options.DaemonOptions) error
	ApplyDashboard(cli kt.CliInterface, options *options.DaemonOptions) error
}

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}
