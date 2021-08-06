package connect

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	return outbound(s, name, podIP, credential, cidrs, cli)
}

func outbound(s *Shadow, podName, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	var stop chan struct{}
	var rootCtx context.Context
	switch s.Options.ConnectOptions.Method {
	case common.ConnectMethodSocks:
		err = forwardSocksTunnelToLocal(s.Options, podName)
	case common.ConnectMethodTun:
		stop, rootCtx, err = forwardSSHTunnelToLocal(s, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startTunConnection(rootCtx, cli, credential, s.Options, podIP, cidrs, stop)
		}
	case common.ConnectMethodSocks5:
		stop, rootCtx, err = forwardSSHTunnelToLocal(s, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startSocks5Connection(cli.Channel(), s.Options)
		}
	default:
		stop, rootCtx, err = forwardSSHTunnelToLocal(s, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			err = startVPNConnection(rootCtx, cli, SSHVPNRequest{
				RemoteSSHHost:          credential.RemoteHost,
				RemoteSSHPKPath:        credential.PrivateKeyPath,
				RemoteSSHPort:          s.Options.ConnectOptions.SSHPort,
				RemoteDNSServerAddress: podIP,
				DisableDNS:             s.Options.ConnectOptions.DisableDNS,
				CustomCRID:             cidrs,
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
