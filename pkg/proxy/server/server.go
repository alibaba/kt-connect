package server

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/proxy/daemon"
	"github.com/rs/zerolog/log"
)

// Run start kt proxy
func Run() (err error) {
	log.Info().Msg("Start kt connect proxy")
	srv := daemon.NewDNSServerDefault()
	err = srv.ListenAndServe()
	if err != nil {
		return
	}
	fmt.Printf("DNS Server Start At 53...\n")
	return
}
