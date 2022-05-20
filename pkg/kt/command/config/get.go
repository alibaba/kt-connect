package config

import (
	"fmt"
)

func Get(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specifiy a config item")
	}
	customConfig := loadCustomConfig()
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config unset --all'")
	}
	for _, item := range args {
		v, err2 := getConfigValue(config, item)
		if err2 != nil {
			return fmt.Errorf("config item '%s' is invalid, please check available config items with 'ktctl config show --all'", item)
		}
		if v != "" {
			fmt.Printf("%s = %v\n", item, v)
		} else if cv, _ := getConfigValue(customConfig, item); cv != "" {
			fmt.Printf("%s = %v  (build-in)\n", item, cv)
		} else {
			fmt.Printf("%s = <empty>\n", item)
		}
	}
	return nil
}

func getConfigValue(config map[string]map[string]string, key string) (string, error) {
	group, item, err := parseConfigItem(key)
	if err != nil {
		return "", err
	}
	if groupValue, exits := config[group]; exits {
		if itemValue, exits2 := groupValue[item]; exits2 {
			return itemValue, nil
		}
	}
	return "", nil
}
