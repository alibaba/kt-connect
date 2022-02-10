package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/sshuttle"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

func BySshuttle() error {
	checkSshuttleInstalled()

	podIP, podName, privateKeyPath, err := getOrCreateShadow()
	if err != nil {
		return err
	}

	cidrs, err := cluster.Ins().ClusterCidrs(context.TODO(), opt.Get().Namespace)
	if err != nil {
		return err
	}

	localSshPort, err := util.GetRandomSSHPort()
	if err != nil {
		return err
	}
	stop, rootCtx, err := transmission.ForwardSSHTunnelToLocal(podName, localSshPort)
	if err != nil {
		return err
	}

	if err = startVPNConnection(rootCtx, &sshuttle.SSHVPNRequest{
		LocalSshPort:           localSshPort,
		RemoteSSHHost:          common.Localhost,
		RemoteSSHPKPath:        privateKeyPath,
		RemoteDNSServerAddress: podIP,
		CustomCIDR:             cidrs,
		Stop:                   stop,
		Debug:                  opt.Get().Debug,
	}); err != nil {
		return err
	}

	return setupDns(podIP)
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
	return util.BackgroundRun(&util.CMDContext{
		Ctx:  rootCtx,
		Cmd:  sshuttle.Ins().Connect(req),
		Name: "vpn(sshuttle)",
		Stop: req.Stop,
	})
}
