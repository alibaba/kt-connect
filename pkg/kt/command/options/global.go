package options

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func GlobalFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:       "Namespace",
			Name:         "namespace",
			Alias:        "n",
			DefaultValue: "",
			Description:  "Specify target namespace (otherwise follow kubeconfig current context)",
		},
		{
			Target:       "KubeConfig",
			Name:         "kubeconfig",
			Alias:        "c",
			DefaultValue: "",
			Description:  "Specify path of KubeConfig file",
		},
		{
			Target:       "Image",
			Name:         "image",
			Alias:        "i",
			DefaultValue: fmt.Sprintf("%s:v%s", util.ImageKtShadow, Store.Version),
			Description:  "Customize shadow image",
		},
		{
			Target:       "ImagePullSecret",
			Name:         "imagePullSecret",
			DefaultValue: "",
			Description:  "Custom image pull secret",
		},
		{
			Target:       "ServiceAccount",
			Name:         "serviceAccount",
			DefaultValue: "default",
			Description:  "Specify ServiceAccount name for shadow pod",
		},
		{
			Target:       "NodeSelector",
			Name:         "nodeSelector",
			DefaultValue: "",
			Description:  "Specify location of shadow and route pod by node label, e.g. 'disk=ssd,region=hangzhou'",
		},
		{
			Target:       "Debug",
			Name:         "debug",
			Alias:        "d",
			DefaultValue: false,
			Description:  "Print debug log",
		},
		{
			Target:       "WithLabels",
			Name:         "withLabel",
			Alias:        "l",
			DefaultValue: "",
			Description:  "Extra labels on shadow pod e.g. 'label1=val1,label2=val2'",
		},
		{
			Target:       "WithAnnotations",
			Name:         "withAnnotation",
			DefaultValue: "",
			Description:  "Extra annotation on shadow pod e.g. 'annotation1=val1,annotation2=val2'",
		},
		{
			Target:       "PortForwardWaitTime",
			Name:         "portForwardTimeout",
			DefaultValue: 10,
			Description:  "Seconds to wait before port-forward connection timeout",
		},
		{
			Target:       "PodCreationWaitTime",
			Name:         "podCreationTimeout",
			DefaultValue: 60,
			Description:  "Seconds to wait before shadow or router pod creation timeout",
		},
		{
			Target:       "UseShadowDeployment",
			Name:         "useShadowDeployment",
			DefaultValue: false,
			Description:  "Deploy shadow container as deployment",
		},
		{
			Target:       "SkipTimeDiff",
			Name:         "useLocalTime",
			DefaultValue: false,
			Description:  "Use local time for resource heartbeat timestamp",
		},
		{
			Target:       "AlwaysUpdateShadow",
			Name:         "forceUpdate",
			Alias:        "f",
			DefaultValue: false,
			Description:  "Always re-pull the latest shadow and router image",
		},
		{
			Target:       "KubeContext",
			Name:         "context",
			DefaultValue: "",
			Description:  "Specify current context of kubeconfig",
		},
		{
			Target:       "PodQuota",
			Name:         "podQuota",
			DefaultValue: "",
			Description:  "Specify resource limit for shadow and router pod, e.g. '0.5c,512m'",
		},
		{
			Target:       "RunAsWorkerProcess",
			Name:         "asWorker",
			DefaultValue: false,
			Description:  "Run as worker process",
			Hidden:       true,
		},
	}
	return flags
}
