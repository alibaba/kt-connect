package command

import (
	"errors"
	"flag"
	fakeKt "github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
	"io/ioutil"
	"testing"
)

func Test_runCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := fakeKt.NewMockCliInterface(ctl)

	mockAction := NewMockActionInterface(ctl)
	mockAction.EXPECT().Preview(gomock.Eq("service"), fakeKtCli, gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"preview", "service", "--expose", "8080", "--external"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"preview", "service"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("--expose is required")},
		{testArgs: []string{"preview"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("an service name must be specified")},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions("test")
		opts.Debug = true
		command := NewPreviewCommand(fakeKtCli, opts, mockAction)
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

