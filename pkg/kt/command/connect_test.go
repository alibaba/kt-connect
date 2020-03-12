package command

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"

	"github.com/alibaba/kt-connect/pkg/fake/kt"
	"github.com/alibaba/kt-connect/pkg/fake/kt/action"
	fakeCluster "github.com/alibaba/kt-connect/pkg/fake/kt/cluster"
	fakeConnect "github.com/alibaba/kt-connect/pkg/fake/kt/connect"
)

func Test_newConnectCommand(t *testing.T) {
	ctl := gomock.NewController(t)

	fakeKtCli := kt.NewMockCliInterface(ctl)

	mockAction := action.NewMockActionInterface(ctl)
	mockAction.EXPECT().Connect(gomock.Any()).Return(nil).AnyTimes()

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
	kubernetes := fakeCluster.NewMockKubernetesInterface(ctl)
	shadow := fakeConnect.NewMockShadowInterface(ctl)
	kubernetes.EXPECT().CreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("172.168.0.2", "shadowName", "sshcm", nil, nil).AnyTimes()
	kubernetes.EXPECT().ClusterCrids(gomock.Any()).Return([]string{"10.10.10.0/24"}, nil)

	shadow.EXPECT().Outbound("shadowName", "172.168.0.2", gomock.Any(), []string{"10.10.10.0/24"}).Return(nil)

	type args struct {
		shadow     connect.ShadowInterface
		kubernetes cluster.KubernetesInterface
		options    *options.DaemonOptions
	}

	opts := options.NewDaemonOptions()
	opts.Labels = "a:b"

	arg := args{
		shadow:     shadow,
		kubernetes: kubernetes,
		options:    opts,
	}

	if err := connectToCluster(arg.shadow, arg.kubernetes, arg.options); (err != nil) != false {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, false)
	}

}

func Test_shouldConnectClusterFailWhenFailCreateShadow(t *testing.T) {

	ctl := gomock.NewController(t)
	kubernetesInterface := fakeCluster.NewMockKubernetesInterface(ctl)
	shadowInterface := fakeConnect.NewMockShadowInterface(ctl)
	kubernetesInterface.EXPECT().CreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", "", nil, errors.New("")).AnyTimes()

	type args struct {
		shadow     connect.ShadowInterface
		kubernetes cluster.KubernetesInterface
		options    *options.DaemonOptions
	}

	arg := args{
		shadow:     shadowInterface,
		kubernetes: kubernetesInterface,
		options:    options.NewDaemonOptions(),
	}

	if err := connectToCluster(arg.shadow, arg.kubernetes, arg.options); (err != nil) != true {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, true)
	}

}

func Test_shouldConnectClusterFailWhenFailGetCrids(t *testing.T) {

	ctl := gomock.NewController(t)
	kubernetes := fakeCluster.NewMockKubernetesInterface(ctl)
	shadow := fakeConnect.NewMockShadowInterface(ctl)
	kubernetes.EXPECT().CreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("172.168.0.2", "shadowName", "sshcm", nil, nil).AnyTimes()
	kubernetes.EXPECT().ClusterCrids(gomock.Any()).Return([]string{}, errors.New("fail to get crid"))

	type args struct {
		shadow     connect.ShadowInterface
		kubernetes cluster.KubernetesInterface
		options    *options.DaemonOptions
	}

	opts := options.NewDaemonOptions()
	opts.Labels = "a:b"

	arg := args{
		shadow:     shadow,
		kubernetes: kubernetes,
		options:    opts,
	}

	if err := connectToCluster(arg.shadow, arg.kubernetes, arg.options); (err != nil) != true {
		t.Errorf("connectToCluster() error = %v, wantErr %v", err, true)
	}

}
