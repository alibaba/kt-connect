package main

import (
	_ "embed"
	"os"
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

func main() {
	tmpl, err := template.New("route").Parse(routeTemplate)
	if err != nil {
		log.Error().Msgf("Failed to load route template")
		return
	}

	data := map[string]interface{}{
		"Port":     "8080",
		"Service":  "tomcat",
		"Versions": []string{"v2", "v3", "v4"},
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		log.Error().Msgf("Failed to generate route configuration")
	}
}
