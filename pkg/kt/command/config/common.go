package config

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

func configFile() string {
	return fmt.Sprintf("%s/config", util.KtHome)
}

func profileFile(profile string) string {
	return fmt.Sprintf("%s/profile/%s", util.KtHome, profile)
}

func loadConfig() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	data, err := ioutil.ReadFile(configFile())
	if err != nil {
		log.Debug().Msgf("Failed to read config file: %s", err)
		if os.IsNotExist(err) {
			return m, nil
		}
		return m, err
	}
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		log.Debug().Msgf("Invalid config content: %s", err)
		return m, err
	}
	return m, nil
}

func saveConfig(config map[string]interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile(), data, 0644)
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