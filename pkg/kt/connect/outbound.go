package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshuttle"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(podName, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	var stop chan struct{}
	var rootCtx context.Context
	switch s.Options.ConnectOptions.Method {
	case common.ConnectMethodSocks:
		err = forwardSocksTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName)
	case common.ConnectMethodTun:
		stop, rootCtx, err = forwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startTunConnection(rootCtx, cli, credential, s.Options, podIP, cidrs, stop)
		}
	case common.ConnectMethodSocks5:
		_, _, err = forwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startSocks5Connection(cli.SshChannel(), s.Options)
		}
	default:
		stop, rootCtx, err = forwardSSHTunnelToLocal(cli.PortForward(), cli.Kubectl(), s.Options, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startVPNConnection(rootCtx, cli, s.Options.ConnectOptions, &sshuttle.SSHVPNRequest{
				RemoteSSHHost:          credential.RemoteHost,
				RemoteSSHPKPath:        credential.PrivateKeyPath,
				RemoteDNSServerAddress: podIP,
				CustomCIDR:             cidrs,
				Stop:                   stop,
				Debug:                  s.Options.Debug,
			})
		}
	}
	if err != nil {
		return
	}

	log.Info().Msgf("Proxy start successful")
	return
}
