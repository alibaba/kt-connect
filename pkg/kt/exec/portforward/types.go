package portforward

import (
	"context"
)

// CliInterface ...
type CliInterface interface {
	ForwardPodPortToLocal(request Request) (chan struct{}, context.Context, error)
}

// Cli ...
type Cli struct{}
