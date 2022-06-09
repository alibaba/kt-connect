package options

func BirdseyeFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:      "SortBy",
			DefaultValue: "name",
			Description: "Sort service by 'name' or 'create-time'",
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
