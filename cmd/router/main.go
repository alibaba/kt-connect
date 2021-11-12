package main

import (
	"github.com/alibaba/kt-connect/pkg/router"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

const actionSetup = "setup"
const actionAdd = "add"
const actionRemove = "remove"

func main() {
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
			log.Error().Msgf("Invalid action action", os.Args[1])
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
	ktConf := router.KtConf{
		Service:  args[0],
		Port:     args[1],
		Versions: []string{args[2]},
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
	err := updateRoute(args[0], actionAdd)
	if err != nil {
		log.Error().Err(err).Msgf("Update route with add failed")
		return
	}
	log.Info().Msgf("Route updated.")
}

func remove(args []string) {
	err := updateRoute(args[0], actionRemove)
	if err != nil {
		log.Error().Err(err).Msgf("Update route with remove failed" )
		return
	}
	log.Info().Msgf("Route updated.")
}

func updateRoute(version string, action string) error {
	ktConf, err := router.ReadKtConf()
	if err != nil {
		return err
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
