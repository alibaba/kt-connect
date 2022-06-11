package options

func BirdseyeFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:      "SortBy",
			DefaultValue: "status",
			Description: "Sort service by 'status' or 'name'",
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
