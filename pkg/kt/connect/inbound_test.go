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

	args := args{
		exposePort: "8080",
		podName:    "podName",
		remoteIP:   "127.0.0.1",
		credential: &util.SSHCredential{},
		options:    options.NewDaemonOptions(),
	}
	args.options.WaitTime = 0

	t.Run("shouldRedirectRequestToLocalHost", func(t *testing.T) {
		kubectl := kubectl.NewMockCliInterface(ctl)
		sshChannel := channel.NewMockChannel(ctl)

		kubectl.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(exec.Command("echo", "kubectl port-forward"))
		sshChannel.EXPECT().ForwardRemoteToLocal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(nil)

		if err := inbound(args.exposePort, args.podName, args.remoteIP, args.credential, args.options, kubectl, sshChannel); err != nil {
			t.Errorf("expect no error, actual is %v", err)
		}
	})
}

type args struct {
	exposePort string
	podName    string
	remoteIP   string
	credential *util.SSHCredential
	options    *options.DaemonOptions
}
