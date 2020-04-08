package command

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	fakeKt "github.com/alibaba/kt-connect/fake/kt"
	"github.com/alibaba/kt-connect/fake/kt/action"
	"github.com/alibaba/kt-connect/fake/kt/cluster"
	"github.com/alibaba/kt-connect/fake/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	coreV1 "k8s.io/api/core/v1"
)

func Test_runCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := fakeKt.NewMockCliInterface(ctl)

	mockAction := action.NewMockActionInterface(ctl)
	mockAction.EXPECT().Run(gomock.Eq("service"), fakeKtCli, gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"run", "service", "--port", "8080", "--expose"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"run", "service"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("--port is required")},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions()
		opts.Debug = true
		command := newRunCommand(fakeKtCli, opts, mockAction)
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

func Test_run(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := fakeKt.NewMockCliInterface(ctl)
	kubernetes := cluster.NewMockKubernetesInterface(ctl)
	shadow := connect.NewMockShadowInterface(ctl)

	fakeKtCli.EXPECT().Kubernetes().AnyTimes().Return(kubernetes, nil)
	fakeKtCli.EXPECT().Shadow().AnyTimes().Return(shadow)

	type args struct {
		service         string
		options         *options.DaemonOptions
		shadowResponse  createShadowResponse
		serviceResponse createServiceResponse
		inboundResponse inboundResponse
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "shouldExposeLocalServiceToCluster",
			args: args{
				service: "test",
				options: options.NewRunDaemonOptions(
					"aa=bb",
					&options.RunOptions{
						Expose: true,
						Port:   8081,
					}),
				shadowResponse: createShadowResponse{
					podIP:   "172.168.0.1",
					podName: "shadow",
					sshcm:   "shadow-ssh-cm",
					credential: &util.SSHCredential{
						RemoteHost:     "127.0.0.1",
						Port:           "2222",
						PrivateKeyPath: "/tmp/pk",
					},
					err: nil,
				},
				serviceResponse: createServiceResponse{
					service: &coreV1.Service{},
					err:     nil,
				},
				inboundResponse: inboundResponse{
					err: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "shouldExposeLocalServiceFailWhenShadowCreateFail",
			args: args{
				service: "test2",
				options: options.NewRunDaemonOptions(
					"aaa=bbb",
					&options.RunOptions{
						Expose: true,
						Port:   8081,
					}),
				shadowResponse: createShadowResponse{
					podIP:   "172.168.0.1",
					podName: "shadow",
					sshcm:   "shadow-ssh-cm",
					credential: &util.SSHCredential{
						RemoteHost:     "127.0.0.1",
						Port:           "2222",
						PrivateKeyPath: "/tmp/pk",
					},
					err: errors.New("fail create shadow"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubernetes.EXPECT().
				GetOrCreateShadow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), false).Times(1).
				Return(tt.args.shadowResponse.podIP, tt.args.shadowResponse.podName, tt.args.shadowResponse.sshcm, tt.args.shadowResponse.credential, tt.args.shadowResponse.err)
			kubernetes.EXPECT().CreateService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(tt.args.serviceResponse.service, tt.args.serviceResponse.err)
			shadow.EXPECT().
				Inbound(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
				Return(tt.args.inboundResponse.err)

			if err := run(tt.args.service, fakeKtCli, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type inboundResponse struct {
	err error
}

type createServiceResponse struct {
	service *coreV1.Service
	err     error
}

type createShadowResponse struct {
	podIP      string
	podName    string
	sshcm      string
	credential *util.SSHCredential
	err        error
}
