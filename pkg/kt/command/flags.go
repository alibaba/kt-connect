package command

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// AppFlags return app flags
func AppFlags(options *options.DaemonOptions, version string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Usage:       "Specify target namespace",
			Value:       "default",
			Destination: &options.Namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Usage:       "Specify path of KubeConfig file",
			Value:       util.KubeConfig(),
			Destination: &options.KubeConfig,
		},
		cli.StringFlag{
			Name:        "serviceAccount",
			Usage:       "Specify ServiceAccount name for shadow pod",
			Value:       "default",
			Destination: &options.ServiceAccount,
		},
		cli.StringFlag{
			Name:        "image,i",
			Usage:       "Custom proxy image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:v" + version,
			Destination: &options.Image,
		},
		cli.StringFlag{
			Name:        "imagePullSecret,s",
			Usage:       "Custom image pull secret",
			Value:       "",
			Destination: &options.ImagePullSecret,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "print debug log",
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
			Value:       10,
		},
		cli.BoolFlag{
			Name:        "forceUpdate,f",
			Usage:       "always update shadow image",
			Destination: &options.ForceUpdateShadow,
		},
		cli.BoolFlag{
			Name:        "useKubectl",
			Usage:       "use kubectl for port-forward",
			Destination: &options.UseKubectl,
		},
	}
}

// ConnectActionFlag ...
func ConnectActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "global",
			Usage:       "With cluster scope",
			Destination: &options.ConnectOptions.Global,
		},
		cli.StringFlag{
			Name:        "method",
			Value:       methodDefaultValue(),
			Usage:       methodDefaultUsage(),
			Destination: &options.ConnectOptions.Method,
		},
		cli.IntFlag{
			Name:        "proxyPort",
			Value:       2223,
			Usage:       "When should method socks5, you can choice which port to proxy",
			Destination: &options.ConnectOptions.SocksPort,
		},
		cli.IntFlag{
			Name:        "sshPort",
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
		cli.StringSliceFlag{
			Name:  "dump2hosts",
			Usage: "Specify namespaces to dump service into local hosts file, use ',' separated",
			Value: &options.ConnectOptions.Dump2HostsNamespaces,
		},
		cli.BoolFlag{
			Name:        "shareShadow",
			Usage:       "Multi clients try to use existing shadow (Beta)",
			Destination: &options.ConnectOptions.ShareShadow,
		},
		cli.StringFlag{
			Name:        "tunName",
			Usage:       "The tun device name to create on client machine (Alpha)",
			Value:       "tun0",
			Destination: &options.ConnectOptions.TunName,
		},
		cli.StringFlag{
			Name:        "tunCidr",
			Usage:       "The cidr used by local tun and peer tun device, at least 4 ips. This cidr MUST NOT overlay with kubernetes service cidr and pod cidr",
			Value:       "10.1.1.0/30",
			Destination: &options.ConnectOptions.TunCidr,
		},
		cli.StringFlag{
			Name:        "clusterDomain",
			Usage:       "The cluster domain provided to kubernetes api-server",
			Value:       "cluster.local",
			Destination: &options.ConnectOptions.ClusterDomain,
		},
		cli.StringFlag{
			Name:        "jvmrc",
			Usage:       "Generate .jvmrc file to specified folder",
			Destination: &options.ConnectOptions.JvmrcDir,
		},
	}
}

func methodDefaultValue() string {
	if util.IsWindows() {
		return common.ConnectMethodSocks
	}
	return common.ConnectMethodVpn
}

func methodDefaultUsage() string {
	if util.IsWindows() {
		return "Connect method 'socks' or 'socks5'"
	} else if util.IsLinux() {
		return "Connect method 'vpn', 'socks', 'socks5' or 'tun'"
	}
	return "Connect method 'vpn', 'socks' or 'socks5'"
}
