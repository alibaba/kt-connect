package tun

import (
	"github.com/linfan/tun2socks/v2/engine"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

// ToSocks create a tun and connect to socks endpoint
func (s *Cli) ToSocks(sockAddr string) error {
	tunSignal := make(chan error)
	go func() {
		var key = new(engine.Key)
		key.Proxy = sockAddr
		key.Device = s.getTunName()
		key.LogLevel = "debug"
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
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}()
	return <-tunSignal
}
