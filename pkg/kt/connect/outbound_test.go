package connect

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"os/exec"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/channel"

	fakeExec "github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func Test_shouldConnectToClusterWithSocks5Methods(t *testing.T) {

	execCli, sshuttle, kubectl, sshChannel := getHandlers(t)

	sshChannel.EXPECT().StartSocks5Proxy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	kubectl.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(namespace, resource, remotePort, localPort interface{}) *exec.Cmd {
			return exec.Command("echo", "kubectl port-forward")
		})
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
	if err := outbound(s, args.name, args.podIP, args.credential, args.cidrs, execCli, sshChannel); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

func Test_shouldConnectToClusterWithVpnMethods(t *testing.T) {

	execCli, sshuttle, kubectl, sshChannel := getHandlers(t)

	kubectl.EXPECT().PortForward(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(namespace, resource, remotePort, localPort interface{}) *exec.Cmd {
			return exec.Command("echo", "kubectl port-forward")
		})
	sshuttle.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(exec.Command("echo", "sshuttle conect"))
	execCli.EXPECT().Kubectl().AnyTimes().Return(kubectl)
	execCli.EXPECT().SSHUttle().AnyTimes().Return(sshuttle)

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
	if err := outbound(s, args.name, args.podIP, args.credential, args.cidrs, execCli, sshChannel); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

func getHandlers(t *testing.T) (*fakeExec.MockCliInterface, *sshuttle.MockCliInterface, *kubectl.MockCliInterface, *channel.MockChannel) {
	ctl := gomock.NewController(t)
	execCli := fakeExec.NewMockCliInterface(ctl)
	sshuttle := sshuttle.NewMockCliInterface(ctl)
	kubectl := kubectl.NewMockCliInterface(ctl)
	sshChannel := channel.NewMockChannel(ctl)
	return execCli, sshuttle, kubectl, sshChannel
}

type OutboundArgs struct {
	name       string
	podIP      string
	credential *util.SSHCredential
	cidrs      []string
}
