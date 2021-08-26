package router

import (
	_ "embed"
	"fmt"
	"os"
	"syscall"
	"text/template"
)

//go:embed route.conf
var routeTemplate string

const pathRouteConf = "/etc/nginx/conf.d/route.conf"

func WriteAndReloadRouteConf(ktConf *KtConf) error {
	var err error
	if len(ktConf.Versions) > 0 {
		err = writeRouteConf(ktConf)
	} else {
		err = removeRouteConf()
	}
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

func writeRouteConf(ktConf *KtConf) error {
	tmpl, err := template.New("route").Parse(routeTemplate)
	if err != nil {
		return fmt.Errorf("failed to load route template: %s", err)
	}

	_ = os.Remove(pathRouteConf)
	routeConfFile, err := os.Create(pathRouteConf)
	if err != nil {
		return fmt.Errorf("failed to create route configuration file: %s", err)
	}
	defer routeConfFile.Close()

	err = tmpl.Execute(routeConfFile, ktConf)
	if err != nil {
		return fmt.Errorf("failed to generate route configuration: %s", err)
	}
	return nil
}

func removeRouteConf() error {
	err := os.Remove(pathRouteConf)
	if err != nil {
		return fmt.Errorf("failed to remove route configuration: %s", err)
	}
	return nil
}
