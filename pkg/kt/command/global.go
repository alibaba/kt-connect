package command

import (
	"path/filepath"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/urfave/cli"
)

// AppFlags return app flags
func AppFlags(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Value:       "default",
			Destination: &options.Namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Value:       filepath.Join(options.RuntimeOptions.UserHome, ".kube", "config"),
			Destination: &options.KubeConfig,
		},
		cli.StringFlag{
			Name:        "image,i",
			Usage:       "Custom proxy image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:latest",
			Destination: &options.Image,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "debug mode",
			Destination: &options.Debug,
		},
		cli.StringFlag{
			Name:        "label,l",
			Usage:       "Extra labels on proxy pod e.g. 'label1=val1,label2=val2'",
			Destination: &options.Labels,
		},
	}
}
