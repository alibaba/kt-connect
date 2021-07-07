package command

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

func Test_cleanCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := kt.NewMockCliInterface(ctl)
	mockAction := NewMockActionInterface(ctl)

	mockAction.EXPECT().Clean(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"clean", "--dryRun"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"clean"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions()
		opts.Debug = true
		command := newCleanCommand(fakeKtCli, opts, mockAction)
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
