package command

import (
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// ActionInterface all action defined
type ActionInterface interface {
	OpenDashboard(options *options.DaemonOptions) error
	Connect(options *options.DaemonOptions) error
	Check(cli kt.CliInterface) error
	Run(service string, options *options.DaemonOptions) error
	Exchange(service string, options *options.DaemonOptions) error
	Mesh(service string, options *options.DaemonOptions) error
	ApplyDashboard(options *options.DaemonOptions) error
}

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}
