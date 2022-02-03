package command

import (
	"errors"
	"flag"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt"

	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

func Test_meshCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeKtCli := kt.NewMockCliInterface(ctl)
	mockAction := NewMockActionInterface(ctl)

	mockAction.EXPECT().Mesh(gomock.Eq("service"), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"mesh", "service", "--expose", "8080"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
		{testArgs: []string{"mesh", "service"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("--expose is required")},
		{testArgs: []string{"mesh"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: errors.New("name of deployment to mesh is required")},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opt.Get().Debug = true
		command := NewMeshCommand(fakeKtCli, mockAction)
		err := command.Run(context)

		if c.expectedErr != nil {
			require.Equal(t, err.Error(), c.expectedErr.Error(), "expected %v but is %v", c.expectedErr, err)
		} else {
			require.Equal(t, err, c.expectedErr, "expected %v but is %v", c.expectedErr, err)
		}
	}
}
