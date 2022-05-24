package profile

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var overwrite bool

func Save(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("must specifiy a profile name")
	}
	profile := profileFile(args[0])
	if _, err := os.Stat(profile); err == nil && !overwrite {
		return fmt.Errorf("profile '%s' already exists, '--overwrite' option is required", args[0])
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

func SaveHandle(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&overwrite, "force", false, "Save even profile with the same name already exists")
}