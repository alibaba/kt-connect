package config

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/spf13/cobra"
	"reflect"
)

var showAll bool

var hiddenOptions = []string{
	"global.as-worker",
}

func Show(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("parameter '%s' is invalid", args[0])
	}
	customConfig := loadCustomConfig()
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config file is damaged, please try repair it or use 'ktctl config unset --all'")
	}
	for i := 0; i < reflect.TypeOf(opt.DaemonOptions{}).NumField(); i++ {
		group := reflect.TypeOf(opt.DaemonOptions{}).Field(i)
		groupName := util.DashSeparated(group.Name)
		for j := 0; j < group.Type.Elem().NumField(); j ++ {
			item := group.Type.Elem().Field(j)
			itemName := util.DashSeparated(item.Name)
			if util.Contains(hiddenOptions, fmt.Sprintf("%s.%s", groupName, itemName)) {
				continue
			}
			if groupValue, groupExist := config[groupName]; groupExist {
				if itemValue, itemExist := groupValue[itemName]; itemExist {
					fmt.Printf("%s.%s = %v\n", groupName, itemName, itemValue)
					continue
				}
			}
			if groupValue, groupExist := customConfig[groupName]; groupExist {
				if itemValue, itemExist := groupValue[itemName]; itemExist {
					fmt.Printf("%s.%s = %v  (build-in)\n", groupName, itemName, itemValue)
					continue
				}
			}
			if showAll {
				fmt.Printf("%s.%s\n", groupName, itemName)
			}
		}
	}
	return nil
}

func ShowHandle(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all available config options")
}
