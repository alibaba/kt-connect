package config

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var profileNamePattern, _ = regexp.Compile("^[a-zA-Z0-9-_.]+$")

func profileFile(profile string) string {
	return fmt.Sprintf("%s/%s", util.KtProfileDir, profile)
}

func loadCustomConfig() map[string]map[string]string {
	config := make(map[string]map[string]string)
	if customize, exist := opt.GetCustomizeKtConfig(); exist {
		_ = yaml.Unmarshal([]byte(customize), &config)
	}
	return config
}

func loadConfig() (map[string]map[string]string, error) {
	config := make(map[string]map[string]string)
	data, err := ioutil.ReadFile(util.KtConfigFile)
	if err != nil {
		log.Debug().Msgf("Failed to read config file: %s", err)
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Debug().Msgf("Invalid config content: %s", err)
		return config, err
	}
	return config, nil
}

func saveConfig(config map[string]map[string]string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(util.KtConfigFile, data, 0644)
}

func parseConfigItem(key string) (string, string, error) {
	if strings.Count(key, ".") != 1 {
		return "", "", fmt.Errorf("config item '%s' is invalid", key)
	}
	parts := strings.Split(key, ".")
	group, exist := reflect.TypeOf(opt.DaemonOptions{}).FieldByName(util.Capitalize(parts[0]))
	if !exist {
		return "", "", fmt.Errorf("config group '%s' not exist", parts[0])
	}
	_, exist = group.Type.Elem().FieldByName(util.Capitalize(parts[1]))
	if !exist {
		return "", "", fmt.Errorf("config item '%s' not exist in group '%s'", parts[1], parts[0])
	}
	return parts[0], parts[1], nil
}

func travelConfigItem(callback func(string, string)) {
	for i := 0; i < reflect.TypeOf(opt.DaemonOptions{}).NumField(); i++ {
		group := reflect.TypeOf(opt.DaemonOptions{}).Field(i)
		groupName := util.DashSeparated(group.Name)
		for j := 0; j < group.Type.Elem().NumField(); j ++ {
			item := group.Type.Elem().Field(j)
			itemName := util.DashSeparated(item.Name)
			callback(groupName, itemName)
		}
	}
}

func configValidator(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var items []string
	if content, err := loadConfig(); err == nil {
		for t, kv := range content {
			for k, _ := range kv {
				items = append(items, t + "." + k)
			}
		}
	}
	return items, cobra.ShellCompDirectiveNoFileComp
}

func profileValidator(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var profiles []string
	if files, err := ioutil.ReadDir(util.KtProfileDir); err == nil {
		for _, f := range files {
			if profileNamePattern.MatchString(f.Name()) {
				profiles = append(profiles, f.Name())
			}
		}
	}
	return profiles, cobra.ShellCompDirectiveNoFileComp
}
