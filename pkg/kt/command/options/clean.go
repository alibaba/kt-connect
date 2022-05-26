package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func CleanFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "ThresholdInMinus",
			DefaultValue: util.ResourceHeartBeatIntervalMinus * 2 + 1,
			Description:  "Length of allowed disconnection time before a unavailing shadow pod be deleted",
		},
		{
			Target:       "DryRun",
			DefaultValue: false,
			Description:  "Only print name of resources to be deleted",
		},
		{
			Target:       "LocalOnly",
			DefaultValue: false,
			Description:  "Only check and restore local changes made by kt",
		},
	}
	return flags
}
