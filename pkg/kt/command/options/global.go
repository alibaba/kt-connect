package options

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func GlobalFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Namespace",
			Alias:        "n",
			DefaultValue: util.DefaultNamespace,
			Description:  "Specify target namespace (otherwise follow kubeconfig current context)",
		},
		{
			Target:       "Kubeconfig",
			Alias:        "c",
			DefaultValue: "",
			Description:  "Specify path of KubeConfig file",
		},
		{
			Target:       "Context",
			DefaultValue: "",
			Description:  "Specify current context of kubeconfig",
		},
		{
			Target:       "Image",
			Alias:        "i",
			DefaultValue: fmt.Sprintf("%s:v%s", util.ImageKtShadow, Store.Version),
			Description:  "Customize shadow image",
		},
		{
			Target:       "ImagePullSecret",
			DefaultValue: "",
			Description:  "Custom image pull secret",
		},
		{
			Target:       "ServiceAccount",
			DefaultValue: "default",
			Description:  "Specify ServiceAccount name for shadow pod",
		},
		{
			Target:       "NodeSelector",
			DefaultValue: "",
			Description:  "Specify location of shadow and route pod by node label, e.g. 'disk=ssd,region=hangzhou'",
		},
		{
			Target:       "Debug",
			Alias:        "d",
			DefaultValue: false,
			Description:  "Print debug log",
		},
		{
			Target:       "WithLabel",
			Alias:        "l",
			DefaultValue: "",
			Description:  "Extra labels on shadow pod e.g. 'label1=val1,label2=val2'",
		},
		{
			Target:       "WithAnnotation",
			DefaultValue: "",
			Description:  "Extra annotation on shadow pod e.g. 'annotation1=val1,annotation2=val2'",
		},
		{
			Target:       "PortForwardTimeout",
			DefaultValue: 10,
			Description:  "Seconds to wait before port-forward connection timeout",
		},
		{
			Target:       "PodCreationTimeout",
			DefaultValue: 60,
			Description:  "Seconds to wait before shadow or router pod creation timeout",
		},
		{
			Target:       "UseShadowDeployment",
			DefaultValue: false,
			Description:  "Deploy shadow container as deployment",
		},
		{
			Target:       "UseLocalTime",
			DefaultValue: false,
			Description:  "Use local time for resource heartbeat timestamp",
		},
		{
			Target:       "ForceUpdate",
			Alias:        "f",
			DefaultValue: false,
			Description:  "Always re-pull the latest shadow and router image",
		},
		{
			Target:       "AsWorker",
			DefaultValue: false,
			Description:  "Run as worker process",
			Hidden:       true,
		},
		{
			Target:       "PodQuota",
			DefaultValue: "",
			Description:  "Specify resource limit for shadow and router pod, e.g. '0.5c,512m'",
		},
	}
	return flags
}
