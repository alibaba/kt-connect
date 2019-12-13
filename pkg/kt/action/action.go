package action

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// Action cmd action
type Action struct {
	Kubeconfig string
	Namespace  string
	Debug      bool
	Image      string
	PidFile    string
	UserHome   string
	Labels     string
	Options *options.DaemonOptions
}
