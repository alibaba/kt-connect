package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"text/template"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

//go:embed route.conf
var routeTemplate string

const pathKtLock = "/var/kt.lock"
const pathKtConf = "/etc/kt.conf"
const pathRouteConf = "/etc/nginx/conf.d/route.conf"
const fieldService = "Service"
const fieldPort = "Port"
const fieldVersions = "Versions"
const actionSetup = "setup"
const actionAdd = "add"
const actionRemove = "remove"

func main() {
	if len(os.Args) < 2 {
		usage()
	} else {
		switch os.Args[0] {
		case actionSetup:
			setup(os.Args[1:])
		case actionAdd:
			add(os.Args[1:])
		case actionRemove:
			remove(os.Args[1:])
		default:
			usage()
		}
	}
}

func usage() {
	log.Error().Msgf(`Usage: 
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
	content := map[string]interface{}{
		fieldService:  args[0],
		fieldPort:     args[1],
		fieldVersions: []string{args[2]},
	}
	err := writeKtConf(content)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	err = writeAndReloadRouteConf(content)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Info().Msgf("Route setup completed.")
}

func add(args []string) {
	err := updateRoute(args[0], actionAdd)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Info().Msgf("Route updated.")
}

func remove(args []string) {
	err := updateRoute(args[0], actionRemove)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	log.Info().Msgf("Route updated.")
}

func updateRoute(version string, action string) error {
	content, err := readKtConf()
	if err != nil {
		return err
	}
	switch action {
	case actionAdd:
		content[fieldVersions] = append(content[fieldVersions].([]string), version)
	case actionRemove:
		versions := content[fieldVersions].([]string)
		for i, v := range versions {
			if v == version {
				content[fieldVersions] = append(versions[:i], versions[i+1:]...)
				break
			}
		}
	}
	err = writeAndReloadRouteConf(content)
	if err != nil {
		return err
	}
	return nil
}

func writeAndReloadRouteConf(content map[string]interface{}) error {
	err := writeRouteConf(content)
	if err != nil {
		return err
	}
	err = reloadRouteConf()
	if err != nil {
		return err
	}
	return nil
}

func reloadRouteConf() error {
	process, err := os.FindProcess(1)
	if err != nil {
		return fmt.Errorf("failed to find route process: %s", err)
	}
	err = process.Signal(syscall.SIGHUP)
	if err != nil {
		return fmt.Errorf("failed to reload route configuration: %s", err)
	}
	return nil
}

func writeRouteConf(content map[string]interface{}) error {
	tmpl, err := template.New("route").Parse(routeTemplate)
	if err != nil {
		return fmt.Errorf("failed to load route template: %s", err)
	}

	_ = os.Remove(pathRouteConf)
	routeConf, err := os.Create(pathRouteConf)
	if err != nil {
		return fmt.Errorf("failed to create route configuration file: %s", err)
	}
	defer routeConf.Close()

	err = tmpl.Execute(routeConf, content)
	if err != nil {
		return fmt.Errorf("failed to generate route configuration: %s", err)
	}
	return nil
}

func readKtConf() (map[string]interface{}, error) {
	ktConf, err := ioutil.ReadFile(pathKtConf)
	if err != nil {
		return nil, fmt.Errorf("failed to read kt configuration file: %s", err)
	}
	var content map[string]interface{}
	err = json.Unmarshal(ktConf, &content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kt configuration file: %s", err)
	}
	return content, nil
}

func writeKtConf(content map[string]interface{}) error {
	bytes, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to parse setup parameters: %s", err)
	}
	err = ioutil.WriteFile(pathKtConf, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to create kt configuration: %s", err)
	}
	return nil
}
