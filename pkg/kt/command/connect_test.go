package command

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"

	"github.com/alibaba/kt-connect/pkg/kt/connect"

	"github.com/alibaba/kt-connect/pkg/kt/exec"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"

	"github.com/alibaba/kt-connect/pkg/kt"
)

func Test_newConnectCommand(t *testing.T) {
	ctl := gomock.NewController(t)

	fakeKtCli := kt.NewMockCliInterface(ctl)

	mockAction := NewMockActionInterface(ctl)
	mockAction.EXPECT().Connect(fakeKtCli, gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"connect", "--method", "socks5"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"connect"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions()
		opts.Debug = true
		command := newConnectCommand(fakeKtCli, opts, mockAction)
		err := command.Run(context)

		if c.expectedErr != nil {
			if err.Error() != c.expectedErr.Error() {
				t.Errorf("expected %v but is %v", c.expectedErr, err)
			}
		} else if err != c.expectedErr {
			t.Errorf("expected %v but is %v", c.expectedErr, err)
		}

	}
}

func Test_shouldConnectToCluster(t *testing.T) {

	ctl := gomock.NewController(t)

	ktctl := kt.NewMockCliInterface(ctl)

	kubernetes := cluster.NewMockKubernetesInterface(ctl)
	exec := exec.NewMockCliInterface(ctl)
	shadow := connect.NewMockShadowInterface(ctl)
	kubernetes.EXPECT().GetOrCreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Return("172.168.0.2", "shadowName", "sshcm", nil, nil).AnyTimes()
	kubernetes.EXPECT().ClusterCrids(gomock.Any(), gomock.Any()).Return([]string{"10.10.10.0/24"}, nil)

	shadow.EXPECT().Outbound("shadowName", "172.168.0.2", gomock.Any(), []string{"10.10.10.0/24"}, gomock.Any()).Return(nil)
	ktctl.EXPECT().Shadow().AnyTimes().Return(shadow)
	ktctl.EXPECT().Kubernetes().AnyTimes().Return(kubernetes, nil)
	ktctl.EXPECT().Exec().AnyTimes().Return(exec)

	opts := options.NewDaemonOptions()
	opts.Labels = "a:b"

	if err := connectToCluster(ktctl, opts); err != nil {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, false)
	}

}

func Test_shouldConnectClusterFailWhenFailCreateShadow(t *testing.T) {

	ctl := gomock.NewController(t)
	ktctl := kt.NewMockCliInterface(ctl)

	kubernetes := cluster.NewMockKubernetesInterface(ctl)
	shadow := connect.NewMockShadowInterface(ctl)
	kubernetes.EXPECT().GetOrCreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Return("", "", "", nil, errors.New("")).AnyTimes()

	ktctl.EXPECT().Shadow().AnyTimes().Return(shadow)
	ktctl.EXPECT().Kubernetes().AnyTimes().Return(kubernetes, nil)

	if err := connectToCluster(ktctl, options.NewDaemonOptions()); err == nil {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, true)
	}

}

func Test_shouldConnectClusterFailWhenFailGetCrids(t *testing.T) {

	ctl := gomock.NewController(t)

	ktctl := kt.NewMockCliInterface(ctl)

	kubernetes := cluster.NewMockKubernetesInterface(ctl)
	shadow := connect.NewMockShadowInterface(ctl)
	kubernetes.EXPECT().GetOrCreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Return("172.168.0.2", "shadowName", "sshcm", nil, nil).AnyTimes()
	kubernetes.EXPECT().ClusterCrids(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("fail to get crid"))

	ktctl.EXPECT().Shadow().AnyTimes().Return(shadow)
	ktctl.EXPECT().Kubernetes().AnyTimes().Return(kubernetes, nil)

	opts := options.NewDaemonOptions()
	opts.Labels = "a:b"

	if err := connectToCluster(ktctl, opts); err == nil {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, true)
	}

}
