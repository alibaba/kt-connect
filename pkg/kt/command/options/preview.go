package options

func PreviewFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Expose",
			DefaultValue: "",
			Description:  "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Required:     true,
		},
		{
			Target:       "External",
			DefaultValue: false,
			Description:  "If specified, a public, external service is created",
		},
		{
			Target:       "SkipPortChecking",
			DefaultValue: false,
			Description:  "Do not check whether specified local ports are listened",
		},
		{
			Target:       "PortNamePrefix",
			DefaultValue: "kt-",
			Description:  "Customize the port name prefix",
		},
	}
	return flags
}
