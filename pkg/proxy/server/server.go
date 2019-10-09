package server

import (
	"github.com/alibaba/kt-connect/pkg/proxy/daemon"
	"github.com/rs/zerolog/log"
)

// Run start kt proxy
func Run() (err error) {
	log.Info().Msg("Start kt connect proxy")
	err = daemon.ServerSSH()
	if err != nil {
		return
	}
	err = daemon.ServeDNS()
	if err != nil {
		return
	}
	return
}
