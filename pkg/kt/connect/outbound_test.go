package connect

import (
	"os/exec"
	"testing"

	fakeExec "github.com/alibaba/kt-connect/pkg/fake/kt/exec"
	"github.com/alibaba/kt-connect/pkg/fake/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/fake/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/fake/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func TestShadow_Outbound(t *testing.T) {

	ctl := gomock.NewController(t)
	execCli := fakeExec.NewMockCliInterface(ctl)

	ssh := ssh.NewMockCliInterface(ctl)
	sshuttle := sshuttle.NewMockCliInterface(ctl)
	kubectl := kubectl.NewMockCliInterface(ctl)

	kubectl.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "kubectl portforward"))
	sshuttle.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "sshuttle conect"))
	ssh.EXPECT().DynamicForwardLocalRequestToRemote(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "ssh dynamic forward"))

	execCli.EXPECT().SSH().AnyTimes().Return(ssh)
	execCli.EXPECT().Kubectl().AnyTimes().Return(kubectl)
	execCli.EXPECT().SSHUttle().AnyTimes().Return(sshuttle)

	type fields struct {
		Options *options.DaemonOptions
	}
	type args struct {
		name       string
		podIP      string
		credential *util.SSHCredential
		cidrs      []string
	}
	vpnOptions := options.NewDaemonOptions()
	socksOptions := options.NewDaemonOptions()
	socksOptions.ConnectOptions.Method = "socks5"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "shouldConnectToClusterWithSocks5Methods",
			fields: fields{
				Options: socksOptions,
			},
			args: args{
				name:  "name",
				podIP: "172.168.0.2",
				credential: &util.SSHCredential{
					RemoteHost:     "127.0.0.1",
					Port:           "223",
					PrivateKeyPath: "/tmp/path",
				},
				cidrs: []string{},
			},
			wantErr: false,
		},
		{
			name: "shouldConnectToClusterWithVpnMethods",
			fields: fields{
				Options: vpnOptions,
			},
			args: args{
				name:  "name",
				podIP: "172.168.0.2",
				credential: &util.SSHCredential{
					RemoteHost:     "127.0.0.1",
					Port:           "223",
					PrivateKeyPath: "/tmp/path",
				},
				cidrs: []string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Shadow{
				Options: tt.fields.Options,
			}
			if err := s.Outbound(tt.args.name, tt.args.podIP, tt.args.credential, tt.args.cidrs, execCli); (err != nil) != tt.wantErr {
				t.Errorf("Shadow.Outbound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
