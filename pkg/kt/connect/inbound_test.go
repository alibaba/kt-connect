package connect

import (
	"os/exec"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/channel"

	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/golang/mock/gomock"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func Test_inbound(t *testing.T) {

	ctl := gomock.NewController(t)

	type args struct {
		exposePort string
		podName    string
		remoteIP   string
		credential *util.SSHCredential
		options    *options.DaemonOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "shouldRedirectRequestToLocalHost",
			args: args{
				exposePort: "8080",
				podName:    "podName",
				remoteIP:   "127.0.0.1",
				credential: &util.SSHCredential{},
				options:    options.NewDaemonOptions(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			kubectl := kubectl.NewMockCliInterface(ctl)
			ssh := ssh.NewMockCliInterface(ctl)

			sshChannel := channel.NewMockChannel(ctl)

			kubectl.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any()).Return(exec.Command("ls", "-al"))
			ssh.EXPECT().ForwardRemoteRequestToLocal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(exec.Command("ls", "-al"))
			sshChannel.EXPECT().ForwardRemoteToLocal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			if err := inbound(tt.args.exposePort, tt.args.podName, tt.args.remoteIP, tt.args.credential, tt.args.options, kubectl, sshChannel); (err != nil) != tt.wantErr {
				t.Errorf("inbound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
