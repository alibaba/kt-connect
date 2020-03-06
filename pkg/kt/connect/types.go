package connect

import "github.com/alibaba/kt-connect/pkg/kt/options"

// ShadowInterface shadow interface
type ShadowInterface interface {
	Inbound(exposePort, podName, remoteIP string) (err error)
	Outbound(name, podIP string, cidrs []string) (err error)
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
