package tun

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/xjasonlyu/tun2socks/v2/engine"
	tunLog "github.com/xjasonlyu/tun2socks/v2/log"
	"os"
	"os/signal"
	"syscall"
)

// ToSocks create a tun and connect to socks endpoint
func (s *Cli) ToSocks(sockAddr string) error {
	tunSignal := make(chan error)
	logLevel := "warning"
	if opt.Get().Global.Debug {
		logLevel = "debug"
	}
	go func() {
		var key = new(engine.Key)
		key.Proxy = sockAddr
		key.Device = fmt.Sprintf("tun://%s", s.GetName())
		key.LogLevel = logLevel
		tunLog.SetOutput(util.BackgroundLogger)
		engine.Insert(key)
		tunSignal <-engine.Start()

		defer func() {
			if err := engine.Stop(); err != nil {
				log.Error().Err(err).Msgf("Stop tun device %s failed", key.Device)
			} else {
				log.Info().Msgf("Tun device %s stopped", key.Device)
			}
		}()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
	}()
	return <-tunSignal
}
