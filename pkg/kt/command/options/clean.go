package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func CleanFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "ThresholdInMinus",
			Name:         "thresholdInMinus",
			DefaultValue: util.ResourceHeartBeatIntervalMinus * 2 + 1,
			Description:  "Length of allowed disconnection time before a unavailing shadow pod be deleted",
		},
		{
			Target:       "DryRun",
			Name:         "dryRun",
			DefaultValue: false,
			Description:  "Only print name of resources to be deleted",
		},
		{
			Target:       "SweepLocalRoute",
			Name:         "sweepLocalRoute",
			DefaultValue: false,
			Description:  "Also clean up local route table record created by kt",
		},
	}
	return flags
}
