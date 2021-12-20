package tunnel

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/portforward"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func ForwardSSHTunnelToLocal(cli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string, localSSHPort int) (stop chan struct{}, rootCtx context.Context, err error) {
	if options.UseKubectl {
		err = PortForwardViaKubectl(kubectlCli, options, podName, common.SshPort, localSSHPort)
	} else {
		stop, rootCtx, err = cli.ForwardPodPortToLocal(options, podName, common.SshPort, localSSHPort)
	}
	return stop, rootCtx, err
}

func PortForwardViaKubectl(kubectlCli kubectl.CliInterface, options *options.DaemonOptions, podName string, remotePort, localPort int) error {
	command := kubectlCli.PortForward(options.Namespace, podName, remotePort, localPort)

	// If localSSHPort is in use by another process, return an error.
	ready := util.WaitPortBeReady(1, localPort)
	if ready {
		return fmt.Errorf("127.0.0.1:%d already in use", localPort)
	}

	err := util.BackgroundRun(command, fmt.Sprintf("forward %d to localhost:%d", remotePort, localPort))
	if err == nil {
		if !util.WaitPortBeReady(options.WaitTime, localPort) {
			err = fmt.Errorf("connect to port-forward failed")
		}
		util.SetupPortForwardHeartBeat(localPort)
	}
	return err
}
