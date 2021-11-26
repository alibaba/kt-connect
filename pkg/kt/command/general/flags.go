package general

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
			Usage:       "Specify target namespace (otherwise follow kubeconfig current context)",
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
			Usage:       "Customize proxy image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:v" + version,
			Destination: &options.Image,
		},
		cli.StringFlag{
			Name:        "imagePullSecret",
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
			Name:        "withLabel,l",
			Usage:       "Extra labels on proxy pod e.g. 'label1=val1,label2=val2'",
			Destination: &options.WithLabels,
		},
		cli.StringFlag{
			Name:        "withAnnotation",
			Usage:       "Extra annotation on proxy pod e.g. 'annotation1=val1,annotation2=val2'",
			Destination: &options.WithAnnotations,
		},
		cli.BoolFlag{
			Name:        "useKubectl",
			Usage:       "use kubectl for port-forward",
			Destination: &options.UseKubectl,
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "extra kubectl options (must use with '--useKubectl') e.g. -e '--cluster=name' -e '--insecure-skip-tls-verify=true'",
			Value: &options.KubeOptions,
		},
		cli.IntFlag{
			Name:        "waitTime",
			Usage:       "seconds to wait before kubectl port-forward connect timeout",
			Destination: &options.WaitTime,
			Value:       10,
		},
		cli.BoolFlag{
			Name:        "forceUpdate,f",
			Usage:       "always update shadow image",
			Destination: &options.AlwaysUpdateShadow,
		},
		cli.StringFlag{
			Name:        "context",
			Usage:       "Specify current context of kubeconfig",
			Destination: &options.KubeContext,
		},
	}
}

// ConnectActionFlag ...
func ConnectActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "method",
			Usage:       methodDefaultUsage(),
			Destination: &options.ConnectOptions.Method,
			Value:       methodDefaultValue(),
		},
		cli.BoolFlag{
			Name:        "shareShadow",
			Usage:       "Use shared shadow pod with other clients (Beta)",
			Destination: &options.ConnectOptions.ShareShadow,
		},
		cli.IntFlag{
			Name:        "sshPort",
			Usage:       "Specify the local port used for SSH port-forward to shadow pod",
			Destination: &options.ConnectOptions.SSHPort,
			Value:       2222,
		},
		cli.BoolFlag{
			Name:        "disablePodIp",
			Usage:       "(vpn mode only) Disable access to pod IP address",
			Destination: &options.ConnectOptions.DisablePodIp,
		},
		cli.StringFlag{
			Name:        "excludeIps",
			Usage:       "(vpn mode only) Do not route specified IPs to cluster, e.g. '192.168.64.2' or '192.168.64.0/24', use ',' separated",
			Destination: &options.ConnectOptions.ExcludeIps,
		},
		cli.StringFlag{
			Name:        "cidr",
			Usage:       "(vpn mode only) Custom CIDR, e.g. '172.2.0.0/16', use ',' separated",
			Destination: &options.ConnectOptions.CIDRs,
		},
		cli.BoolFlag{
			Name:        "disableDNS",
			Usage:       "(vpn / tun mode only) Disable Cluster DNS",
			Destination: &options.ConnectOptions.DisableDNS,
		},
		cli.StringFlag{
			Name:        "tunName",
			Usage:       "(tun method only) The tun device name to create on client machine (Alpha)",
			Destination: &options.ConnectOptions.TunName,
			Value:       "tun0",
		},
		cli.StringFlag{
			Name:        "tunCidr",
			Usage:       "(tun method only) The cidr used by local tun and peer tun device, at least 4 ips. This cidr MUST NOT overlay with kubernetes service cidr and pod cidr",
			Destination: &options.ConnectOptions.TunCidr,
			Value:       "10.1.1.0/30",
		},
		cli.StringFlag{
			Name:        "proxyAddr",
			Usage:       "(socks5 method only) Specify the address which socks5 proxy should listen to",
			Destination: &options.ConnectOptions.SocksAddr,
			Value:       "127.0.0.1",
		},
		cli.IntFlag{
			Name:        "proxyPort",
			Usage:       "(socks5 method only) Specify the local port which socks5 proxy should use",
			Destination: &options.ConnectOptions.SocksPort,
			Value:       2223,
		},
		cli.StringFlag{
			Name:        "dump2hosts",
			Usage:       "(socks / socks5 method only) Specify namespaces to dump service into local hosts file, use ',' separated",
			Destination: &options.ConnectOptions.Dump2HostsNamespaces,
		},
		cli.StringFlag{
			Name:        "clusterDomain",
			Usage:       "(socks / socks5 method only) The cluster domain provided to kubernetes api-server",
			Destination: &options.ConnectOptions.ClusterDomain,
			Value:       "cluster.local",
		},
		cli.StringFlag{
			Name:        "jvmrc",
			Usage:       "(Windows only) Generate .jvmrc file to specified folder",
			Destination: &options.ConnectOptions.JvmrcDir,
		},
		cli.BoolFlag{
			Name:        "setupGlobalProxy",
			Usage:       "(Windows only) Auto setup HTTP_PROXY variable and system global proxy configuration",
			Destination: &options.ConnectOptions.UseGlobalProxy,
		},
	}
}

func ExchangeActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "expose",
			Usage:       "ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.ExchangeOptions.Expose,
		},
		cli.IntFlag{
			Name:        "recoverWaitTime",
			Usage:       "seconds to wait for original deployment recover before turn off the shadow pod",
			Destination: &options.ExchangeOptions.RecoverWaitTime,
			Value:       120,
		},
		cli.StringFlag{
			Name:        "method",
			Usage:       "Exchange method 'scale' or 'ephemeral'(experimental)",
			Destination: &options.ExchangeOptions.Method,
			Value:       "scale",
		},
	}
}

func MeshActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "expose",
			Usage:       "ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.MeshOptions.Expose,
		},
		cli.StringFlag{
			Name:        "versionMark",
			Usage:       "specify the version of mesh service, e.g. '0.0.1' or 'kt-version:local'",
			Destination: &options.MeshOptions.VersionMark,
		},
		cli.StringFlag{
			Name:        "method",
			Value:       "manual",
			Usage:       "Mesh method 'manual' or 'auto'(beta)",
			Destination: &options.MeshOptions.Method,
		},
		cli.StringFlag{
			Name:        "routerImage",
			Usage:       "(auto method only) Customize proxy image",
			Destination: &options.MeshOptions.RouterImage,
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:v" + options.RuntimeOptions.Version,
		},
	}
}

func ProvideActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:        "expose",
			Usage:       "The port that exposes",
			Destination: &options.ProvideOptions.Expose,
		},
		cli.BoolFlag{
			Name:        "external",
			Usage:       "If specified, a public, external service is created",
			Destination: &options.ProvideOptions.External,
		},
	}
}

func CleanActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "dryRun",
			Usage:       "Only print name of deployments to be deleted",
			Destination: &options.CleanOptions.DryRun,
		},
		cli.Int64Flag{
			Name:        "thresholdInMinus",
			Usage:       "Length of allowed disconnection time before a unavailing shadow pod be deleted",
			Destination: &options.CleanOptions.ThresholdInMinus,
			Value:       util.ResourceHeartBeatIntervalMinus * 3,
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
