package config

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)

func ListProfile(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("parameter '%s' is invalid", args[0])
	}
	files, err := ioutil.ReadDir(util.KtProfileDir)
	if err != nil {
		log.Error().Msgf("Failed to list profiles")
		return err
	}
	fmt.Println("Save Date          \t\tName")
	for _, f := range files {
		if !f.IsDir() {
			fmt.Printf("%s\t\t%s\n", f.ModTime().Format(common.YyyyMmDdHhMmSs), f.Name())
		}
	}
	return nil
}

