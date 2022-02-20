package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/sshuttle"
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

	cidrs, err := cluster.Ins().ClusterCidrs(opt.Get().Namespace)
	if err != nil {
		return err
	}

	localSshPort, err := util.GetRandomTcpPort()
	if err != nil {
		return err
	}
	stop, err := transmission.SetupPortForwardToLocal(podName, common.StandardSshPort, localSshPort)
	if err != nil {
		return err
	}

	if err = startVPNConnection(&sshuttle.SSHVPNRequest{
		LocalSshPort:           localSshPort,
		RemoteSSHPKPath:        privateKeyPath,
		RemoteDNSServerAddress: podIP,
		CustomCIDR:             cidrs,
	}, stop); err != nil {
		return err
	}

	return setupDns(podName, podIP)
}

func checkSshuttleInstalled() {
	if !util.CanRun(sshuttle.Ins().Version()) {
		_, _, err := util.RunAndWait(sshuttle.Ins().Install())
		if err != nil {
			log.Error().Err(err).Msgf("Failed find or install sshuttle")
		}
	}
}

func startVPNConnection(req *sshuttle.SSHVPNRequest, stop chan struct{}) (err error) {
	return util.BackgroundRun(&util.CMDContext{
		Ctx:  context.Background(),
		Cmd:  sshuttle.Ins().Connect(req),
		Name: "vpn(sshuttle)",
		Stop: stop,
	})
}
