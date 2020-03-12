package connect

import (
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// ShadowInterface shadow interface
type ShadowInterface interface {
	Inbound(exposePort, podName, remoteIP string, credential *util.SSHCredential) (err error)
	Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, exec exec.CliInterface) (err error)
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
