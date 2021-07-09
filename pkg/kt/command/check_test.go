package command

import (
	"flag"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/options"

	fakeExec "github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/ssh"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"

	fakeKt "github.com/alibaba/kt-connect/pkg/kt"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

func TestNewCheckCommand(t *testing.T) {

	ctl := gomock.NewController(t)
	fakeCli := fakeKt.NewMockCliInterface(ctl)
	fakeAction := NewMockActionInterface(ctl)

	fakeAction.EXPECT().Check(fakeCli).Return(nil)

	cases := []struct {
		testArgs               []string
		skipFlagParsing        bool
		useShortOptionHandling bool
		expectedErr            error
	}{
		{testArgs: []string{"check"}, skipFlagParsing: false, useShortOptionHandling: false, expectedErr: nil},
	}

	for _, c := range cases {

		app := &cli.App{Writer: ioutil.Discard}
		set := flag.NewFlagSet("test", 0)
		_ = set.Parse(c.testArgs)

		context := cli.NewContext(app, set, nil)

		opts := options.NewDaemonOptions()
		opts.Debug = true
		command := newCheckCommand(fakeCli, opts, fakeAction)
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

func TestAction_CheckSuccessful(t *testing.T) {
	// given
	ctl := gomock.NewController(t)
	sshCli := ssh.NewMockCliInterface(ctl)
	sshuttleCli := sshuttle.NewMockCliInterface(ctl)
	kubectlCli := kubectl.NewMockCliInterface(ctl)

	sshCli.EXPECT().Version().AnyTimes().Return(exec.Command("echo", "ssh 0.0.1"))
	sshuttleCli.EXPECT().Version().AnyTimes().Return(exec.Command("echo", "sshuttle 0.0.2"))
	kubectlCli.EXPECT().Version().AnyTimes().Return(exec.Command("echo", "kubectl 0.0.1"))

	exec := fakeExec.NewMockCliInterface(ctl)
	fakeCli := fakeKt.NewMockCliInterface(ctl)

	exec.EXPECT().SSH().Return(sshCli)
	exec.EXPECT().SSHUttle().Return(sshuttleCli)
	exec.EXPECT().Kubectl().Return(kubectlCli)

	fakeCli.EXPECT().Exec().AnyTimes().Return(exec)

	action := &Action{Options: options.NewDaemonOptions()}

	// when
	if err := action.Check(fakeCli); (err != nil) != false {
		t.Errorf("Action.Check() error = %v, wantErr %v", err, false)
	}
}
