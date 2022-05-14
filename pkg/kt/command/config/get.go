package config

import (
	"fmt"
)

func Get(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specifiy a config item")
	}
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config reset --all'")
	}
	for _, item := range args {
		v, err2 := getConfigValue(config, item)
		if err2 != nil {
			return fmt.Errorf("config item '%s' is invalid, please check available config items with 'ktctl config show'", item)
		}
		if v != nil {
			fmt.Printf("%s = %v\n", item, v)
		} else {
			fmt.Printf("%s not defined\n", item)
		}
	}
	return nil
}

func getConfigValue(config map[interface{}]interface{}, key string) (interface{}, error) {
	group, item, err := parseConfigItem(key)
	if err != nil {
		return nil, err
	}
	if groupValue, exits := config[group]; exits {
		if itemValue, exits2 := groupValue.(map[string]interface{})[item]; exits2 {
			return itemValue, nil
		}
	}
	return nil, nil
}
