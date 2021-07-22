package connect

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/channel"
	"github.com/alibaba/kt-connect/pkg/kt/exec"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
)

// Outbound start vpn connection
func (s *Shadow) Outbound(name, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface) (err error) {
	ssh := channel.SSHChannel{}
	return outbound(s, name, podIP, credential, cidrs, cli, &ssh)
}

func outbound(s *Shadow, podName, podIP string, credential *util.SSHCredential, cidrs []string, cli exec.CliInterface, ssh channel.Channel) (err error) {
	if s.Options.ConnectOptions.Method == common.ConnectMethodSocks {
		err = forwardSocksTunnelToLocal(s.Options, podName)
	} else {
		stop, rootCtx, err := forwardSSHTunnelToLocal(s, podName, s.Options.ConnectOptions.SSHPort)
		if err == nil {
			if s.Options.ConnectOptions.Method == common.ConnectMethodSocks5 {
				err = startSocks5Connection(ssh, s.Options)
			} else {
				err = startVPNConnection(rootCtx, cli, credential, s.Options, podIP, cidrs, stop)
			}
		}
	}
	if err != nil {
		return
	}

	log.Info().Msgf("KT proxy start successful")
	return
}
