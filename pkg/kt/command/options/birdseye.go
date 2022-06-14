package options

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func BirdseyeFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:      "SortBy",
			DefaultValue: util.SortByStatus,
			Description: fmt.Sprintf("Sort service by '%s' or '%s'", util.SortByStatus, util.SortByName),
		},
		{
			Target:      "ShowConnector",
			DefaultValue: false,
			Description: "Also show name of users who connected to cluster",
		},
		{
			Target:      "HideNaturalService",
			DefaultValue: false,
			Description: "Only show exchanged / meshed and previewing services",
		},
	}
	return flags
}
