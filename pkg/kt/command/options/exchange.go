package options

import "github.com/alibaba/kt-connect/pkg/kt/util"

func ExchangeFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Expose",
			Name:         "expose",
			DefaultValue: "",
			Description:  "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Required:     true,
		},
		{
			Target:       "Mode",
			Name:         "mode",
			DefaultValue: util.ExchangeModeSelector,
			Description:  "Exchange method 'selector', 'scale' or 'ephemeral'(experimental)",
		},
		{
			Target:       "RecoverWaitTime",
			Name:         "recoverWaitTime",
			DefaultValue: 120,
			Description:  "(scale method only) Seconds to wait for original deployment recover before turn off the shadow pod",
		},
	}
	return flags
}
