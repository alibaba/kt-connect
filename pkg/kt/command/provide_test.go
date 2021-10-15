package command

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"

	"github.com/alibaba/kt-connect/pkg/kt/connect"

	fakeKt "github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	coreV1 "k8s.io/api/core/v1"
)

func Test_runCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := fakeKt.NewMockCliInterface(ctl)

	mockAction := NewMockActionInterface(ctl)
	mockAction.EXPECT().Provide(gomock.Eq("service"), fakeKtCli, gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"provide", "service", "--expose", "8080", "--external"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"provide", "service"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("--expose is required")},
		{testArgs: []string{"provide"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("an service name must be specified")},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions("test")
		opts.Debug = true
		command := newProvideCommand(fakeKtCli, opts, mockAction)
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

func Test_shouldExposeLocalServiceToCluster(t *testing.T) {
	fakeKtCli, kubernetes, shadow := getHandlers(t)

	args := args{
		service: "test",
		options: testDaemonOptions(
			"aa=bb",
			&options.ProvideOptions{
				External: false,
				Expose:   8081,
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
	}

	kubernetes.EXPECT().GetOrCreateShadow(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
		Return(args.shadowResponse.podIP, args.shadowResponse.podName, args.shadowResponse.sshcm, args.shadowResponse.credential, args.shadowResponse.err)
	kubernetes.EXPECT().CreateService(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(args.serviceResponse.service, args.serviceResponse.err)
	shadow.EXPECT().Inbound(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(args.inboundResponse.err)

	if err := provide(context.TODO(), args.service, fakeKtCli, args.options); err != nil {
		t.Errorf("expect no error, actual is %v", err)
	}
}

func Test_shouldExposeLocalServiceFailWhenShadowCreateFail(t *testing.T) {
	fakeKtCli, kubernetes, shadow := getHandlers(t)

	args := args{
		service: "test2",
		options: testDaemonOptions(
			"aaa=bbb",
			&options.ProvideOptions{
				External: false,
				Expose:   8081,
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
	}

	kubernetes.EXPECT().GetOrCreateShadow(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
		Return(args.shadowResponse.podIP, args.shadowResponse.podName, args.shadowResponse.sshcm, args.shadowResponse.credential, args.shadowResponse.err)
	kubernetes.EXPECT().CreateService(context.TODO(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(args.serviceResponse.service, args.serviceResponse.err)
	shadow.EXPECT().Inbound(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0).Return(args.inboundResponse.err)

	if err := provide(context.TODO(), args.service, fakeKtCli, args.options); err == nil {
		t.Errorf("expect error = %v, actual is nil", err)
	}
}

func testDaemonOptions(labels string, opt *options.ProvideOptions) *options.DaemonOptions {
	daemonOptions := options.NewDaemonOptions("test")
	daemonOptions.WithLabels = labels
	daemonOptions.ProvideOptions = opt
	return daemonOptions
}

func getHandlers(t *testing.T) (*fakeKt.MockCliInterface, *cluster.MockKubernetesInterface, *connect.MockShadowInterface) {
	ctl := gomock.NewController(t)
	fakeKtCli := fakeKt.NewMockCliInterface(ctl)
	kubernetes := cluster.NewMockKubernetesInterface(ctl)
	shadow := connect.NewMockShadowInterface(ctl)

	fakeKtCli.EXPECT().Kubernetes().AnyTimes().Return(kubernetes, nil)
	fakeKtCli.EXPECT().Shadow().AnyTimes().Return(shadow)
	return fakeKtCli, kubernetes, shadow
}

type args struct {
	service         string
	options         *options.DaemonOptions
	shadowResponse  createShadowResponse
	serviceResponse createServiceResponse
	inboundResponse inboundResponse
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
