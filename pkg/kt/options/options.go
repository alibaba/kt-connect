package options

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"k8s.io/client-go/kubernetes"

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
	SSHPort              int
	SocksPort            int
	CIDR                 string
	Method               string
	Dump2HostsNamespaces cli.StringSlice
	ShareShadow          bool
	LocalDomain          string
	TunName              string
	TunCidr              string
	ClusterDomain        string

	// Used for tun mode
	SourceIP string
	DestIP   string
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Expose string
}

// MeshOptions ...
type MeshOptions struct {
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
	// Path of user home, same as ${HOME}
	UserHome string
	// Path of kt config folder, default to ${UserHome}/.ktctl
	AppHome string
	// Component current sub-command
	Component string
	// Shadow deployment name
	Shadow string
	// ssh public key name of config map. format is kt-xxx(component)-public-key-xxx(version)
	SSHCM string
	// The origin app name
	Origin string
	// The origin replicas
	Replicas int32
	// Exposed service name
	Service string
	// Whether dump2host enabled
	Dump2Host bool
	// Windows global proxy config
	ProxyConfig registry.ProxyConfig
}

type dashboardOptions struct {
	Install bool
	Port    string
}

// DaemonOptions cli options
type DaemonOptions struct {
	KubeConfig        string
	Namespace         string
	Debug             bool
	Image             string
	Labels            string
	KubeOptions       cli.StringSlice
	RuntimeOptions    *RuntimeOptions
	ProvideOptions    *ProvideOptions
	ConnectOptions    *ConnectOptions
	ExchangeOptions   *ExchangeOptions
	MeshOptions       *MeshOptions
	CleanOptions      *CleanOptions
	DashboardOptions  *dashboardOptions
	WaitTime          int
	ForceUpdateShadow bool
}

// NewDaemonOptions return new cli default options
func NewDaemonOptions() *DaemonOptions {
	userHome := util.UserHome
	appHome := util.KtHome
	util.CreateDirIfNotExist(appHome)
	return &DaemonOptions{
		Namespace:  common.DefNamespace,
		KubeConfig: util.KubeConfig(),
		WaitTime:   5,
		RuntimeOptions: &RuntimeOptions{
			UserHome: userHome,
			AppHome:  appHome,
		},
		ConnectOptions:   &ConnectOptions{},
		ExchangeOptions:  &ExchangeOptions{},
		MeshOptions:      &MeshOptions{},
		CleanOptions:     &CleanOptions{},
		DashboardOptions: &dashboardOptions{},
		ProvideOptions:   &ProvideOptions{},
	}
}

// NewProvideDaemonOptions ...
func NewProvideDaemonOptions(labels string, options *ProvideOptions) *DaemonOptions {
	daemonOptions := NewDaemonOptions()
	daemonOptions.Labels = labels
	daemonOptions.ProvideOptions = options
	return daemonOptions
}
