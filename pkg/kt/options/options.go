package options

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// ProvideOptions ...
type ProvideOptions struct {
	External bool
	Expose   int
}

// ConnectOptions ...
type ConnectOptions struct {
	Global               bool
	DisableDNS           bool
	DisablePodIp         bool
	SSHPort              int
	SocksPort            int
	CIDRs                cli.StringSlice
	Method               string
	Dump2HostsNamespaces cli.StringSlice
	ShareShadow          bool
	TunName              string
	TunCidr              string
	ClusterDomain        string
	JvmrcDir             string
	UseGlobalProxy       bool

	// Used for tun mode
	SourceIP string
	DestIP   string
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Method string
	Expose string
	Label  string
}

// ExchangePodOptions ...
type ExchangePodOptions struct {
	Label  string
	Expose string
}

// MeshOptions ...
type MeshOptions struct {
	Method  string
	Expose  string
	Version string
}

// CleanOptions ...
type CleanOptions struct {
	DryRun           bool
	ThresholdInMinus int64
}

// RuntimeOptions ...
type RuntimeOptions struct {
	Clientset kubernetes.Interface
	// UserHome path of user home, same as ${HOME}
	UserHome string
	// AppHome path of kt config folder, default to ${UserHome}/.ktctl
	AppHome string
	// Component current sub-command
	Component string
	// Shadow deployment name
	Shadow string
	// SSHCM ssh public key name of config map. format is kt-xxx(component)-public-key-xxx(version)
	SSHCM string
	// Origin the origin app name
	Origin string
	// Replicas the origin replicas
	Replicas int32
	// Service exposed service name
	Service string
	// Dump2Host whether dump2host enabled
	Dump2Host bool
	// ProxyConfig windows global proxy config
	ProxyConfig registry.ProxyConfig
	// RestConfig kubectl config
	RestConfig *rest.Config
	PodName    string
}

type dashboardOptions struct {
	Install bool
	Port    string
}

// DaemonOptions cli options
type DaemonOptions struct {
	Version            string
	KubeConfig         string
	Namespace          string
	ServiceAccount     string
	Debug              bool
	Image              string
	ImagePullSecret    string
	Labels             string
	KubeOptions        cli.StringSlice
	RuntimeOptions     *RuntimeOptions
	ProvideOptions     *ProvideOptions
	ConnectOptions     *ConnectOptions
	ExchangeOptions    *ExchangeOptions
	ExchangePodOptions *ExchangePodOptions
	MeshOptions        *MeshOptions
	CleanOptions       *CleanOptions
	DashboardOptions   *dashboardOptions
	WaitTime           int
	ForceUpdateShadow  bool
	UseKubectl         bool
}

// NewDaemonOptions return new cli default options
func NewDaemonOptions(version string) *DaemonOptions {
	return &DaemonOptions{
		Namespace:  common.DefNamespace,
		KubeConfig: util.KubeConfig(),
		WaitTime:   5,
		Version:    version,
		RuntimeOptions: &RuntimeOptions{
			UserHome: util.UserHome,
			AppHome:  util.KtHome,
		},
		ConnectOptions:     &ConnectOptions{},
		ExchangeOptions:    &ExchangeOptions{},
		ExchangePodOptions: &ExchangePodOptions{},
		MeshOptions:        &MeshOptions{},
		CleanOptions:       &CleanOptions{},
		DashboardOptions:   &dashboardOptions{},
		ProvideOptions:     &ProvideOptions{},
	}
}
