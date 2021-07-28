package network

import (
	"context"
)

// CliInterface ...
type CliInterface interface {
	PortForward(req PortForwardAPodRequest) error
	ForwardPodPortToLocal(request PortForwardAPodRequest) (chan struct{}, context.Context, error)
}

// Cli ...
type Cli struct{}
