package config

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var dryRun bool

func LoadProfile(args []string) error {
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
	if dryRun {
		fmt.Println(bytesRead)
	} else {
		err = ioutil.WriteFile(util.KtConfigFile, bytesRead, 0644)
		if err != nil {
			return fmt.Errorf("unable to save config file: %s", err)
		}
		log.Info().Msgf("Profile '%s' loaded", args[0])
	}
	return nil
}

func LoadProfileHandle(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&dryRun, "dryRun", false, "Print profile content without load it")
}
