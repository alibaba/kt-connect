package command

import (
	"flag"
	"io/ioutil"
	"testing"

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

		opts := options.NewDaemonOptions("test")
		opts.Debug = true
		command := NewConnectCommand(fakeKtCli, opts, mockAction)
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

func TestAllocateIP(t *testing.T) {
	cidr := "10.1.1.0/30"

	srcIP, destIP, err := allocateTunIP(cidr)
	if err != nil {
		t.Errorf("allocateTunIP() error: %v", err)
	}

	if srcIP != "10.1.1.1" {
		t.Errorf("allocateTunIP() failed, current: %s, want: %s", srcIP, "10.1.1.1")
	}

	if destIP != "10.1.1.2" {
		t.Errorf("allocateTunIP() failed, current: %s, want: %s", destIP, "10.1.1.2")
	}
}
