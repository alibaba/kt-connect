package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// ActionInterface all action defined
type ActionInterface interface {
	OpenDashboard(options *options.DaemonOptions) error
	Connect(options *options.DaemonOptions) error
	Check(options *options.DaemonOptions) error
	Run(service string, options *options.DaemonOptions) error
	Exchange(service string, options *options.DaemonOptions) error
	Mesh(service string, options *options.DaemonOptions) error
	ApplyDashboard() error
}

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}
