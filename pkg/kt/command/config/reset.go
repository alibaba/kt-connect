package config

import (
	"fmt"
	"github.com/spf13/cobra"
)

var unsetAll bool

func Unset(args []string) error {
	if unsetAll {
		return saveConfig(make(map[string]map[string]string))
	}
	if len(args) < 1 {
		return fmt.Errorf("must specifiy a config item")
	}
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config unset --all'")
	}
	for _, item := range args {
		err = unsetConfigValue(config, item)
		if err != nil {
			return fmt.Errorf("%s, please check available config items with 'ktctl config show --all'", err)
		}
	}
	return saveConfig(config)
}

func UnsetHandle(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&unsetAll, "all", "a", false, "Unset all config options")
}

func unsetConfigValue(config map[string]map[string]string, key string) error {
	group, item, err := parseConfigItem(key)
	if err != nil {
		return err
	}
	if _, exist := config[group]; exist {
		delete(config[group], item)
	}
	return nil
}
