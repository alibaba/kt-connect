package main

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/router"
	"github.com/gofrs/flock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

const pathKtLock = "/var/kt.lock"
const actionSetup = "setup"
const actionAdd = "add"
const actionRemove = "remove"

func main() {
	fileLock := flock.New(pathKtLock)
	if err := fileLock.Lock(); err != nil {
		log.Error().Err(err).Msgf("Unable to fetch route lock")
		return
	}
	defer fileLock.Unlock()
	if len(os.Args) < 3 {
		usage()
	} else {
		switch os.Args[1] {
		case actionSetup:
			setup(os.Args[2:])
		case actionAdd:
			add(os.Args[2:])
		case actionRemove:
			remove(os.Args[2:])
		default:
			log.Error().Msgf("Invalid action '%s'", os.Args[1])
			usage()
		}
	}
}

func usage() {
	log.Info().Msgf(`Usage: 
router %s <service-name> <service-port> <custom-version>
router %s <custom-version>
router %s <custom-version>
`, actionSetup, actionAdd, actionRemove)
}

func setup(args []string) {
	if len(args) < 3 {
		usage()
		return
	}
	header, version := splitVersionMark(args[2])
	ktConf := router.KtConf{
		Service:  args[0],
		Ports:    getPorts(args[1]),
		Header:   header,
		Versions: []string{version},
	}
	err := router.WriteKtConf(&ktConf)
	if err != nil {
		log.Error().Err(err).Msgf("Write kt config failed")
		return
	}
	err = router.WriteAndReloadRouteConf(&ktConf)
	if err != nil {
		log.Error().Err(err).Msgf("Write and load route config failed")
		return
	}
	log.Info().Msgf("Route setup completed.")
}

func add(args []string) {
	header, version := splitVersionMark(args[0])
	err := updateRoute(header, version, actionAdd)
	if err != nil {
		log.Error().Err(err).Msgf("Update route with add failed")
		return
	}
	log.Info().Msgf("Route updated.")
}

func remove(args []string) {
	header, version := splitVersionMark(args[0])
	err := updateRoute(header, version, actionRemove)
	if err != nil {
		log.Error().Err(err).Msgf("Update route with remove failed" )
		return
	}
	log.Info().Msgf("Route updated.")
}

func splitVersionMark(mark string) (string, string) {
	splits := strings.Split(mark, ":")
	return strings.ReplaceAll(splits[0], "-", "_"), splits[1]
}

func getPorts(portsParameter string) [][]string {
	ports := make([][]string, 0)
	for _, pp := range strings.Split(portsParameter, ",") {
		ports = append(ports, strings.Split(pp, ":"))
	}
	return ports
}

func updateRoute(header, version, action string) error {
	ktConf, err := router.ReadKtConf()
	if err != nil {
		return err
	}
	if ktConf.Header != header {
		return fmt.Errorf("specified header '%s' no match mesh pod header '%s'", header, ktConf.Header)
	}
	switch action {
	case actionAdd:
		ktConf.Versions = append(ktConf.Versions, version)
	case actionRemove:
		versions := ktConf.Versions
		for i, v := range versions {
			if v == version {
				ktConf.Versions = append(versions[:i], versions[i+1:]...)
				break
			}
		}
	}
	err = router.WriteKtConf(ktConf)
	if err != nil {
		return err
	}
	err = router.WriteAndReloadRouteConf(ktConf)
	if err != nil {
		return err
	}
	return nil
}
