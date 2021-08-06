package network

import (
	"context"
)

// CliInterface ...
type CliInterface interface {
	ForwardPodPortToLocal(request PortForwardAPodRequest) (chan struct{}, context.Context, error)
}

// Cli ...
type Cli struct{}
