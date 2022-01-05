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
		cli.StringFlag{
			Name:        "nodeSelector",
			Usage:       "Specify location of shadow and route pod by node label, e.g. 'disk=ssd,region=hangzhou'",
			Value:       "",
			Destination: &options.NodeSelector,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "Print debug log",
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
		cli.IntFlag{
			Name:        "portForwardTimeout",
			Usage:       "Seconds to wait before port-forward connection timeout",
			Destination: &options.PortForwardWaitTime,
			Value:       10,
		},
		cli.IntFlag{
			Name:        "podCreationTimeout",
			Usage:       "Seconds to wait before shadow or router pod creation timeout",
			Destination: &options.PodCreationWaitTime,
			Value:       60,
		},
		cli.BoolFlag{
			Name:        "forceUpdate,f",
			Usage:       "Always update shadow image",
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
			Name:        "mode",
			Usage:       "Connect mode 'tun2socks' or 'sshuttle'",
			Destination: &options.ConnectOptions.Mode,
			Value:       common.ConnectModeTun2Socks,
		},
		cli.BoolFlag{
			Name:        "shareShadow",
			Usage:       "Use standalone shadow pod",
			Destination: &options.ConnectOptions.SoleShadow,
		},
		cli.IntFlag{
			Name:        "sshPort",
			Usage:       "Specify the local port used for SSH port-forward to shadow pod",
			Destination: &options.ConnectOptions.SSHPort,
			Value:       2222,
		},
		cli.StringFlag{
			Name:        "dumpToHosts",
			Usage:       "Specify namespaces to dump service into local hosts file, use ',' separated",
			Destination: &options.ConnectOptions.Dump2HostsNamespaces,
		},
		cli.StringFlag{
			Name:        "clusterDomain",
			Usage:       "The cluster domain provided to kubernetes api-server",
			Destination: &options.ConnectOptions.ClusterDomain,
			Value:       "cluster.local",
		},
		cli.BoolFlag{
			Name:        "disablePodIp",
			Usage:       "Disable access to pod IP address",
			Destination: &options.ConnectOptions.DisablePodIp,
		},
		cli.BoolFlag{
			Name:        "disableDNS",
			Usage:       "Disable Cluster DNS",
			Destination: &options.ConnectOptions.DisableDNS,
		},
		cli.StringFlag{
			Name:        "includeIps",
			Usage:       "Specify extra IP ranges which should be route to cluster, e.g. '172.2.0.0/16', use ',' separated",
			Destination: &options.ConnectOptions.IncludeIps,
		},
		cli.StringFlag{
			Name:        "excludeIps",
			Usage:       "(sshuttle mode only) Do not route specified IPs to cluster, e.g. '192.168.64.2' or '192.168.64.0/24', use ',' separated",
			Destination: &options.ConnectOptions.ExcludeIps,
		},
		cli.BoolFlag{
			Name:        "disableTunDevice",
			Usage:       "(tun2socks mode only) Create socks5 proxy without tun device",
			Destination: &options.ConnectOptions.DisableTunDevice,
		},
		cli.BoolFlag{
			Name:        "disableTunRoute",
			Usage:       "(tun2socks mode only) Do not auto setup tun device route",
			Destination: &options.ConnectOptions.DisableTunRoute,
		},
		cli.IntFlag{
			Name:        "proxyPort",
			Usage:       "(tun2socks mode only) Specify the local port which socks5 proxy should use",
			Destination: &options.ConnectOptions.SocksPort,
			Value:       2223,
		},
	}
}

func ExchangeActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "mode",
			Usage:       "Exchange method 'switch', 'scale' or 'ephemeral'(experimental)",
			Destination: &options.ExchangeOptions.Mode,
			Value:       common.ExchangeModeSwitch,
		},
		cli.StringFlag{
			Name:        "expose",
			Usage:       "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.ExchangeOptions.Expose,
		},
		cli.IntFlag{
			Name:        "recoverWaitTime",
			Usage:       "(scale method only) Seconds to wait for original deployment recover before turn off the shadow pod",
			Destination: &options.ExchangeOptions.RecoverWaitTime,
			Value:       120,
		},
	}
}

func MeshActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "mode",
			Usage:       "Mesh method 'auto' or 'manual'",
			Destination: &options.MeshOptions.Mode,
			Value:       common.MeshModeAuto,
		},
		cli.StringFlag{
			Name:        "expose",
			Usage:       "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.MeshOptions.Expose,
		},
		cli.StringFlag{
			Name:        "versionMark",
			Usage:       "Specify the version of mesh service, e.g. '0.0.1' or 'kt-version:local'",
			Destination: &options.MeshOptions.VersionMark,
		},
		cli.StringFlag{
			Name:        "routerImage",
			Usage:       "(auto method only) Customize router image",
			Destination: &options.MeshOptions.RouterImage,
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:v" + options.RuntimeOptions.Version,
		},
	}
}

func PreviewActionFlag(options *options.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "expose",
			Usage:       "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.PreviewOptions.Expose,
		},
		cli.BoolFlag{
			Name:        "external",
			Usage:       "If specified, a public, external service is created",
			Destination: &options.PreviewOptions.External,
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
