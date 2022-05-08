package options

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func MeshFlags() []OptionConfig {
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
			DefaultValue: util.MeshModeAuto,
			Description:  "Mesh method 'auto' or 'manual'",
		},
		{
			Target:       "VersionMark",
			Name:         "versionMark",
			DefaultValue: "",
			Description:  "Specify the version of mesh service, e.g. '0.0.1' or 'mark:local'",
		},
		{
			Target:       "RouterImage",
			Name:         "routerImage",
			DefaultValue: fmt.Sprintf("%s:v%s", util.ImageKtRouter, Get().RuntimeStore.Version),
			Description:  "(auto method only) Customize router image",
		},
	}
	return flags
}
