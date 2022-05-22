package config

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

func Load(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("must specifiy a profile name")
	}
	profile := profileFile(args[0])
	if _, err := os.Stat(profile); err != nil {
		return fmt.Errorf("profile '%s' not exists", args[0])
	}
	bytesRead, err := ioutil.ReadFile(profile)
	if err != nil {
		return fmt.Errorf("unable to read profile file: %s", err)
	}
	err = ioutil.WriteFile(util.KtConfigFile, bytesRead, 0644)
	if err != nil {
		return fmt.Errorf("unable to save config file: %s", err)
	}
	log.Info().Msgf("Profile '%s' loaded", args[0])
	return nil
}
