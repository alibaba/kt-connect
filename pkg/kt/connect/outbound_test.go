package connect

import (
	"github.com/alibaba/kt-connect/pkg/common"
	fakeExec "github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/golang/mock/gomock"
	"os/exec"
	"testing"
)

func Test_shouldConnectToClusterWithSocks5Methods(t *testing.T) {

	execCli, sshuttle, kubectl, sshChannel := getHandlers(t)

	sshChannel.EXPECT().StartSocks5Proxy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	sshuttle.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "sshuttle conect"))
	execCli.EXPECT().Kubectl().AnyTimes().Return(kubectl)
	execCli.EXPECT().SSHUttle().AnyTimes().Return(sshuttle)

	socksOptions := options.NewDaemonOptions()
	socksOptions.ConnectOptions.Method = common.ConnectMethodSocks5
	socksOptions.WaitTime = 0

	args := OutboundArgs{
		name:  "name",
		podIP: "172.168.0.2",
		credential: &util.SSHCredential{
			RemoteHost:     "127.0.0.1",
			Port:           "223",
			PrivateKeyPath: "/tmp/path",
		},
		cidrs: []string{},
	}

	s := &Shadow{
		Options: socksOptions,
	}
	if err := outbound(s, args.name, args.podIP, args.credential, args.cidrs, execCli); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

func Test_shouldConnectToClusterWithVpnMethods(t *testing.T) {

	execCli, sshuttle, kubectl, sshChannel := getHandlers(t)

	sshuttle.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "sshuttle conect"))
	execCli.EXPECT().Kubectl().AnyTimes().Return(kubectl)
	execCli.EXPECT().SSHUttle().AnyTimes().Return(sshuttle)
	execCli.EXPECT().Channel().AnyTimes().Return(sshChannel)

	vpnOptions := options.NewDaemonOptions()
	vpnOptions.WaitTime = 0

	args := OutboundArgs{
		name:  "name",
		podIP: "172.168.0.2",
		credential: &util.SSHCredential{
			RemoteHost:     "127.0.0.1",
			Port:           "223",
			PrivateKeyPath: "/tmp/path",
		},
		cidrs: []string{},
	}

	s := &Shadow{
		Options: vpnOptions,
	}
	if err := outbound(s, args.name, args.podIP, args.credential, args.cidrs, execCli); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

func getHandlers(t *testing.T) (*fakeExec.MockCliInterface, *sshuttle.MockCliInterface, *kubectl.MockCliInterface, *sshchannel.MockChannel) {
	ctl := gomock.NewController(t)
	execCli := fakeExec.NewMockCliInterface(ctl)
	sshuttle := sshuttle.NewMockCliInterface(ctl)
	kubectl := kubectl.NewMockCliInterface(ctl)
	sshChannel := sshchannel.NewMockChannel(ctl)
	return execCli, sshuttle, kubectl, sshChannel
}

type OutboundArgs struct {
	name       string
	podIP      string
	credential *util.SSHCredential
	cidrs      []string
}
