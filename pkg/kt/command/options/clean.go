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
			Target:       "SweepLocalRoute",
			DefaultValue: false,
			Description:  "Also clean up local route table record created by kt",
		},
	}
	return flags
}
