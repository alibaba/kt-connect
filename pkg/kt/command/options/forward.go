package options

func ForwardFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Port",
			DefaultValue: "",
			Description:  "Specify local port to listening, default to the same port as the forwarded target",
		},
	}
	return flags
}
