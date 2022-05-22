package config

import (
	"fmt"
	"strings"
)

func Set(args []string) error {
	if len(args) < 1 || len(args) > 2 || (len(args) == 1 && !strings.Contains(args[0], "=")) {
		return fmt.Errorf("please use either 'set <item>=<value>' or 'set <item> <value>' format")
	}
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config unset --all'")
	}
	var key, value string
	if len(args) == 1 {
		parts := strings.SplitN(args[0], "=", 2)
		key = parts[0]
		value = parts[1]
	} else {
		key = args[0]
		value = args[1]
	}
	err = setConfigValue(config, key, value)
	if err != nil {
		return fmt.Errorf("%s, please check available config items with 'ktctl config show --all'", err)
	}
	return saveConfig(config)
}

func setConfigValue(config map[string]map[string]string, key string, value string) error {
	group, item, err := parseConfigItem(key)
	if err != nil {
		return err
	}
	if _, exist := config[group]; exist {
		config[group][item] = value
	} else {
		config[group] = map[string]string { item: value }
	}
	return nil
}
