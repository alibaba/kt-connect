package command

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt"

	"github.com/golang/mock/gomock"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/urfave/cli"
)

func Test_exchangeCommand(t *testing.T) {
	ctl := gomock.NewController(t)
	fakeKtCli := kt.NewMockCliInterface(ctl)
	mockAction := NewMockActionInterface(ctl)

	mockAction.EXPECT().Exchange(gomock.Eq("service"), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"exchange", "service", "--expose", "8080"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"exchange", "service"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("--expose is required")},
		{testArgs: []string{"exchange"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("name of deployment to exchange is required")},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions("test")
		opts.Debug = true
		command := newExchangeCommand(fakeKtCli, opts, mockAction)
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
