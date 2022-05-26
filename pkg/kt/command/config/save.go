package config

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"strings"
)

func SaveProfile(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("must specifiy a profile name")
	}
	profile := profileFile(args[0])
	if _, err := os.Stat(profile); err == nil {
		var answer string
		fmt.Printf("Profile '%s' already exists, overwrite ? (Y/N) ", args[0])
		_, err = fmt.Scanln(&answer)
		if err != nil || !strings.HasPrefix(strings.ToUpper(answer), "Y") {
			return nil
		}
	}
	bytesRead, err := ioutil.ReadFile(util.KtConfigFile)
	if err != nil {
		return fmt.Errorf("unable to read config file: %s", err)
	}
	err = ioutil.WriteFile(profile, bytesRead, 0644)
	if err != nil {
		return fmt.Errorf("unable to create profile file: %s", err)
	}
	log.Info().Msgf("Profile '%s' saved", args[0])
	return nil
}
