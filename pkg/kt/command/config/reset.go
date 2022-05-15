package config

import (
	"fmt"
	"github.com/spf13/cobra"
)

var resetAll bool

func Reset(args []string) error {
	if resetAll {
		return saveConfig(make(map[string]map[string]string))
	}
	if len(args) < 1 {
		return fmt.Errorf("must specifiy a config item")
	}
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config reset --all'")
	}
	for _, item := range args {
		err = resetConfigValue(config, item)
		if err != nil {
			return fmt.Errorf("%s, please check available config items with 'ktctl config show --all'", err)
		}
	}
	return saveConfig(config)
}

func ResetHandle(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&resetAll, "all", false, "Reset all config options")
}

func resetConfigValue(config map[string]map[string]string, key string) error {
	group, item, err := parseConfigItem(key)
	if err != nil {
		return err
	}
	if _, exist := config[group]; exist {
		delete(config[group], item)
	}
	return nil
}
