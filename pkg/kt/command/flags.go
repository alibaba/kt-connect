package command

import (
	"path/filepath"

	"github.com/alibaba/kt-connect/pkg/kt/util"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/urfave/cli"
)

// PluginConnectFlags kubectl-connect flags
func PluginConnectFlags(options *options.DaemonOptions) []cli.Flag {
	globalFlags := AppFlags(options)
	connectFlags := ConnectActionFlag(options)

	flags := []cli.Flag{}
	flags = append(flags, globalFlags...)
	flags = append(flags, connectFlags...)
	return flags
}

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

// ConnectActionFlag ...
func ConnectActionFlag(options *options.DaemonOptions) []cli.Flag {
	methodDefaultValue := "vpn"
	methodDefaultUsage := "Connect method 'vpn' or 'socks5'"
	if util.IsWindows() {
		methodDefaultValue = "socks5"
		methodDefaultUsage = "windows only support socks5"
	}

	return []cli.Flag{
		cli.StringFlag{
			Name:        "method",
			Value:       methodDefaultValue,
			Usage:       methodDefaultUsage,
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
			Usage:       "Custom CIDR eq '172.2.0.0/16'",
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
			Usage:       "Multi clients try to use existing shadow",
			Destination: &options.ConnectOptions.ShareShadow,
		},
	}
}
