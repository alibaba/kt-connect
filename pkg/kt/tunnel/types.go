package tunnel

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// ShadowInterface shadow interface
type ShadowInterface interface {
	Inbound(exposePort, podName string) (int, error)
}

// Shadow shadow
type Shadow struct {
	Options *options.DaemonOptions
}

// Create create shadow
func Create(options *options.DaemonOptions) (shadow Shadow) {
	shadow = Shadow{
		Options: options,
	}
	return
}
