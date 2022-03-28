package general

import (
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// AppFlags return app flags
func AppFlags(options *opt.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "namespace,n",
			Usage:       "Specify target namespace (otherwise follow kubeconfig current context)",
			Destination: &options.Namespace,
		},
		cli.StringFlag{
			Name:        "kubeconfig,c",
			Usage:       "Specify path of KubeConfig file",
			Destination: &options.KubeConfig,
		},
		cli.StringFlag{
			Name:        "image,i",
			Usage:       "Customize shadow image",
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow:v" + options.RuntimeStore.Version,
			Destination: &options.Image,
		},
		cli.StringFlag{
			Name:        "imagePullSecret",
			Usage:       "Custom image pull secret",
			Value:       "",
			Destination: &options.ImagePullSecret,
		},
		cli.StringFlag{
			Name:        "serviceAccount",
			Usage:       "Specify ServiceAccount name for shadow pod",
			Value:       "default",
			Destination: &options.ServiceAccount,
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
			Usage:       "Extra labels on shadow pod e.g. 'label1=val1,label2=val2'",
			Destination: &options.WithLabels,
		},
		cli.StringFlag{
			Name:        "withAnnotation",
			Usage:       "Extra annotation on shadow pod e.g. 'annotation1=val1,annotation2=val2'",
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
			Name:        "useShadowDeployment",
			Usage:       "Deploy shadow container as deployment",
			Destination: &options.UseShadowDeployment,
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
		cli.StringFlag{
			Name:        "podQuota",
			Usage:       "Specify resource limit for shadow and router pod, e.g. '0.5c,512m'",
			Destination: &options.PodQuota,
		},
	}
}

// ConnectActionFlag ...
func ConnectActionFlag(options *opt.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "mode",
			Usage:       "Connect mode 'tun2socks' or 'sshuttle'",
			Destination: &options.ConnectOptions.Mode,
			Value:       util.ConnectModeTun2Socks,
		},
		cli.StringFlag{
			Name:        "dnsMode",
			Usage:       "Specify how to resolve service domains, can be 'localDNS', 'podDNS', 'hosts' or 'hosts:<namespaces>', for multiple namespaces use ',' separation",
			Destination: &options.ConnectOptions.DnsMode,
			Value:       util.DnsModeLocalDns,
		},
		cli.BoolFlag{
			Name:        "shareShadow",
			Usage:       "Use shared shadow pod",
			Destination: &options.ConnectOptions.SharedShadow,
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
			Name:        "skipCleanup",
			Usage:       "Do not auto cleanup residual resources in cluster",
			Destination: &options.ConnectOptions.SkipCleanup,
		},
		cli.StringFlag{
			Name:        "includeIps",
			Usage:       "Specify extra IP ranges which should be route to cluster, e.g. '172.2.0.0/16', use ',' separated",
			Destination: &options.ConnectOptions.IncludeIps,
		},
		cli.StringFlag{
			Name:        "excludeIps",
			Usage:       "Do not route specified IPs to cluster, e.g. '192.168.64.2' or '192.168.64.0/24', use ',' separated",
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
		cli.Int64Flag{
			Name:       "dnsCacheTtl",
			Usage:      "(local dns mode only) DNS cache refresh interval in seconds",
			Destination: &options.ConnectOptions.DnsCacheTtl,
			Value:       60,
		},
	}
}

func ExchangeActionFlag(options *opt.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "mode",
			Usage:       "Exchange method 'selector', 'scale' or 'ephemeral'(experimental)",
			Destination: &options.ExchangeOptions.Mode,
			Value:       util.ExchangeModeSelector,
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

func MeshActionFlag(options *opt.DaemonOptions) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "mode",
			Usage:       "Mesh method 'auto' or 'manual'",
			Destination: &options.MeshOptions.Mode,
			Value:       util.MeshModeAuto,
		},
		cli.StringFlag{
			Name:        "expose",
			Usage:       "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80",
			Destination: &options.MeshOptions.Expose,
		},
		cli.StringFlag{
			Name:        "versionMark",
			Usage:       "Specify the version of mesh service, e.g. '0.0.1' or 'mark:local'",
			Destination: &options.MeshOptions.VersionMark,
		},
		cli.StringFlag{
			Name:        "routerImage",
			Usage:       "(auto method only) Customize router image",
			Destination: &options.MeshOptions.RouterImage,
			Value:       "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:v" + options.RuntimeStore.Version,
		},
	}
}

func PreviewActionFlag(options *opt.DaemonOptions) []cli.Flag {
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

func RecoverActionFlag(options *opt.DaemonOptions) []cli.Flag {
	return []cli.Flag{
	}
}

func CleanActionFlag(options *opt.DaemonOptions) []cli.Flag {
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
			Value:       cluster.ResourceHeartBeatIntervalMinus * 2 + 1,
		},
	}
}
