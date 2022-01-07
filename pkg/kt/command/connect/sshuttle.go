package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

func BySshuttle(cli kt.CliInterface, options *options.DaemonOptions) error {
	if len(options.ConnectOptions.Dump2HostsNamespaces) > 0 {
		options.RuntimeOptions.Dump2Host = setupDump2Host(cli.Kubernetes(), options.Namespace,
			options.ConnectOptions.Dump2HostsNamespaces, options.ConnectOptions.ClusterDomain)
	}
	checkSshuttleInstalled(cli.Exec().Sshuttle(), options.Debug)

	podIP, podName, credential, err := getOrCreateShadow(cli.Kubernetes(), options)
	if err != nil {
		return err
	}

	cidrs, err := cli.Kubernetes().ClusterCidrs(context.TODO(), options.Namespace, options.ConnectOptions)
	if err != nil {
		return err
	}

	stop, rootCtx, err := tunnel.ForwardSSHTunnelToLocal(options, podName, options.ConnectOptions.SSHPort)
	if err != nil {
		return err
	}
	return startVPNConnection(rootCtx, cli.Exec(), options.ConnectOptions, &sshuttle.SSHVPNRequest{
		RemoteSSHHost:          credential.RemoteHost,
		RemoteSSHPKPath:        credential.PrivateKeyPath,
		RemoteDNSServerAddress: podIP,
		CustomCIDR:             cidrs,
		Stop:                   stop,
		Debug:                  options.Debug,
	})
}

func checkSshuttleInstalled(cli sshuttle.CliInterface, isDebug bool) {
	if !util.CanRun(cli.Version()) {
		err := util.RunAndWait(cli.Install(), isDebug)
		if err != nil {
			log.Error().Err(err).Msgf("Failed find or install sshuttle")
		}
	}
}

func startVPNConnection(rootCtx context.Context, cli exec.CliInterface, opt *options.ConnectOptions, req *sshuttle.SSHVPNRequest) (err error) {
	err = util.BackgroundRun(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  cli.Sshuttle().Connect(opt, req),
		Name: "vpn(sshuttle)",
		IsDebug: true,
		Stop: req.Stop,
	})
	return err
}
