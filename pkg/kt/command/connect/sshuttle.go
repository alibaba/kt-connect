package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/kt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

func BySshuttle(cli kt.CliInterface) error {
	checkSshuttleInstalled()

	podIP, podName, credential, err := getOrCreateShadow(cli.Kubernetes())
	if err != nil {
		return err
	}

	cidrs, err := cli.Kubernetes().ClusterCidrs(context.TODO(), opt.Get().Namespace)
	if err != nil {
		return err
	}

	stop, rootCtx, err := transmission.ForwardSSHTunnelToLocal(podName, opt.Get().ConnectOptions.SSHPort)
	if err != nil {
		return err
	}

	if err = startVPNConnection(rootCtx, &sshuttle.SSHVPNRequest{
		RemoteSSHHost:          credential.RemoteHost,
		RemoteSSHPKPath:        credential.PrivateKeyPath,
		RemoteDNSServerAddress: podIP,
		CustomCIDR:             cidrs,
		Stop:                   stop,
		Debug:                  opt.Get().Debug,
	}); err != nil {
		return err
	}

	return setupDns(cli, podIP)
}

func checkSshuttleInstalled() {
	if !util.CanRun(sshuttle.Ins().Version()) {
		_, _, err := util.RunAndWait(sshuttle.Ins().Install())
		if err != nil {
			log.Error().Err(err).Msgf("Failed find or install sshuttle")
		}
	}
}

func startVPNConnection(rootCtx context.Context, req *sshuttle.SSHVPNRequest) (err error) {
	err = util.BackgroundRun(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  sshuttle.Ins().Connect(req),
		Name: "vpn(sshuttle)",
		Stop: req.Stop,
	})
	return err
}
