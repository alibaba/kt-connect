package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func ConnectFlags() []OptionConfig {
	flags := []OptionConfig{
		{
			Target:      "Mode",
			DefaultValue: util.ConnectModeTun2Socks,
			Description: "Connect mode 'tun2socks' or 'sshuttle'",
		},
		{
			Target:      "DnsMode",
			DefaultValue: util.DnsModeLocalDns,
			Description: "Specify how to resolve service domains, can be 'localDNS', 'podDNS', 'hosts' or 'hosts:<namespaces>', for multiple namespaces use ',' separation",
		},
		{
			Target:      "ShareShadow",
			DefaultValue: false,
			Description: "Use shared shadow pod",
		},
		{
			Target:      "ClusterDomain",
			DefaultValue: "cluster.local",
			Description: "The cluster domain provided to kubernetes api-server",
		},
		{
			Target:      "DisablePodIp",
			DefaultValue: false,
			Description: "Disable access to pod IP address",
		},
		{
			Target:      "SkipCleanup",
			DefaultValue: false,
			Description: "Do not auto cleanup residual resources in cluster",
		},
		{
			Target:      "IncludeIps",
			DefaultValue: "",
			Description: "Specify extra IP ranges which should be route to cluster, e.g. '172.2.0.0/16', use ',' separated",
		},
		{
			Target:      "ExcludeIps",
			DefaultValue: "",
			Description: "Do not route specified IPs to cluster, e.g. '192.168.64.2' or '192.168.64.0/24', use ',' separated",
		},
		{
			Target:      "DisableTunDevice",
			DefaultValue: false,
			Description: "(tun2socks mode only) Create socks5 proxy without tun device",
		},
		{
			Target:      "DisableTunRoute",
			DefaultValue: false,
			Description: "(tun2socks mode only) Do not auto setup tun device route",
		},
		{
			Target:      "ProxyPort",
			DefaultValue: 2223,
			Description: "(tun2socks mode only) Specify the local port which socks5 proxy should use",
		},
		{
			Target:      "DnsCacheTtl",
			DefaultValue: 60,
			Description: "(local dns mode only) DNS cache refresh interval in seconds",
		},
	}
	if util.IsMacos() {
		flags = append(flags, OptionConfig{
			Target:      "DnsPort",
			DefaultValue: util.AlternativeDnsPort,
			Description: "(local dns mode only) Specify local DNS port",
		})
	}
	return flags
}
