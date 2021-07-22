package connect

import (
	"os/exec"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/channel"

	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/golang/mock/gomock"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func Test_shouldRedirectRequestToLocalHost(t *testing.T) {

	ctl := gomock.NewController(t)

	args := InboundArgs{
		exposePort: "8080",
		podName:    "podName",
		remoteIP:   "127.0.0.1",
		credential: &util.SSHCredential{},
		options:    options.NewDaemonOptions(),
	}
	args.options.WaitTime = 0

	kubectlCli := kubectl.NewMockCliInterface(ctl)
	sshChannel := channel.NewMockChannel(ctl)

	kubectlCli.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(exec.Command("echo", "kubectl port-forward"))
	sshChannel.EXPECT().ForwardRemoteToLocal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)

	if err := inbound(nil, args.exposePort, args.podName, args.remoteIP, sshChannel); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

type InboundArgs struct {
	exposePort string
	podName    string
	remoteIP   string
	credential *util.SSHCredential
	options    *options.DaemonOptions
}
