package options

func PreviewFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Expose",
			Name:         "expose",
			DefaultValue: "",
			Description:  "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Required:     true,
		},
		{
			Target:       "External",
			Name:         "external",
			DefaultValue: false,
			Description:  "If specified, a public, external service is created",
		},
	}
	return flags
}
