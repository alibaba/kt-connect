package main

import (
	_ "embed"
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

func main() {
	tmpl, err := template.New("route").Parse(routeTemplate)
	if err != nil {
		log.Error().Msgf("Failed to load route template: %s", err)
		return
	}

	data := map[string]interface{}{
		"Port":     "8080",
		"Service":  "tomcat",
		"Versions": []string{"v2", "v3", "v4"},
	}

	routeConf, err := os.Create(pathRouteConf)
	if err != nil {
		log.Error().Msgf("Failed to create route configuration file: %s", err)
		return
	}
	defer routeConf.Close()

	err = tmpl.Execute(routeConf, data)
	if err != nil {
		log.Error().Msgf("Failed to generate route configuration: %s", err)
		return
	}

	process, err := os.FindProcess(1)
	if err != nil {
		log.Error().Msgf("Failed to find route process: %s", err)
		return
	}
	err = process.Signal(syscall.SIGHUP)
	if err != nil {
		log.Error().Msgf("Failed to reload route configuration: %s", err)
	}

	log.Info().Msgf("Route updated.")

}
