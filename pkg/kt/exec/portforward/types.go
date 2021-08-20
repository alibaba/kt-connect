package portforward

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// CliInterface ...
type CliInterface interface {
	ForwardPodPortToLocal(options *options.DaemonOptions, podName string, remotePort, localPort int) (chan struct{}, context.Context, error)
}

// Cli ...
type Cli struct{}
