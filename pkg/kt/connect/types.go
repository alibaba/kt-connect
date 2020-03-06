package connect

import "github.com/alibaba/kt-connect/pkg/kt/options"

// ShadowInterface shadow interface
type ShadowInterface interface {
	Connect(name, podIP string, cidrs []string, options *options.DaemonOptions) (err error)
	RemotePortForward(expose, kubeconfig, namespace, target, remoteIP string, debug bool) (err error)
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
}
