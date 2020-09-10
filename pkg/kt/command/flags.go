package command

import (
	"path/filepath"

	"github.com/alibaba/kt-connect/pkg/kt/util"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/urfave/cli"
)

// AppFlags return app flags
func AppFlags(options *options.DaemonOptions, version string) []cli.Flag {
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
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:" + version,
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
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "support kubectl options e.g. -e '-n default' -e '--context=kubernetes-admin' -e '--kubeconfig=/path/to/kube/config'",
			Value: &options.KubeOptions,
		},
		cli.IntFlag{
			Name:        "waitTime",
			Usage:       "custom wait time for kubectl port-forward to support slow network environment",
			Destination: &options.WaitTime,
			Value:       5,
		},
	}
}

// ConnectActionFlag ...
func ConnectActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "method",
			Value:       methodDefaultValue(),
			Usage:       methodDefaultUsage(),
			Destination: &options.ConnectOptions.Method,
		},
		cli.IntFlag{
			Name:        "proxy",
			Value:       2223,
			Usage:       "when should method socks5, you can choice which port to proxy",
			Destination: &options.ConnectOptions.Socke5Proxy,
		},
		cli.IntFlag{
			Name:        "port",
			Value:       2222,
			Usage:       "Local SSH Proxy port",
			Destination: &options.ConnectOptions.SSHPort,
		},
		cli.BoolFlag{
			Name:        "disableDNS",
			Usage:       "Disable Cluster DNS",
			Destination: &options.ConnectOptions.DisableDNS,
		},
		cli.StringFlag{
			Name:        "cidr",
			Usage:       "Custom CIDR, e.g. '172.2.0.0/16'",
			Destination: &options.ConnectOptions.CIDR,
		},
		cli.BoolFlag{
			Name:        "dump2hosts",
			Usage:       "Auto write service to local hosts file",
			Destination: &options.ConnectOptions.Dump2Hosts,
		},
		cli.StringSliceFlag{
			Name:  "dump2hostsNS",
			Usage: "Which namespaces service to local hosts file, support multiple namespaces.",
			Value: &options.ConnectOptions.Dump2HostsNamespaces,
		},
		cli.BoolFlag{
			Name:        "shareShadow",
			Usage:       "Multi clients try to use existing shadow (Beta)",
			Destination: &options.ConnectOptions.ShareShadow,
		},
		cli.StringFlag{
			Name:        "localDomain",
			Usage:       "Set local domain suffix to help dns resolve properly",
			Destination: &options.ConnectOptions.LocalDomain,
		},
	}
}

func methodDefaultValue() string {
	if util.IsWindows() {
		return "socks5"
	} else {
		return "vpn"
	}
}

func methodDefaultUsage() string {
	if util.IsWindows() {
		return "Windows only support socks5"
	} else {
		return "Connect method 'vpn' or 'socks5'"
	}
}
