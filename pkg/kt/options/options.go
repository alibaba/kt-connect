package options

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// ConnectOptions ...
type ConnectOptions struct {
	Global               bool
	DisableDNS           bool
	DisablePodIp         bool
	SSHPort              int
	SocksPort            int
	SocksAddr            string
	CIDRs                string
	ExcludeIps           string
	Method               string
	Dump2HostsNamespaces string
	ShareShadow          bool
	TunName              string
	TunCidr              string
	ClusterDomain        string
	JvmrcDir             string
	UseGlobalProxy       bool
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Method          string
	Expose          string
	RecoverWaitTime int
}

// MeshOptions ...
type MeshOptions struct {
	Method      string
	Expose      string
	VersionMark string
	RouterImage string
}

// ProvideOptions ...
type ProvideOptions struct {
	External bool
	Expose   string
}

// CleanOptions ...
type CleanOptions struct {
	DryRun           bool
	ThresholdInMinus int64
}

// RuntimeOptions ...
type RuntimeOptions struct {
	Clientset kubernetes.Interface
	// Version ktctl version
	Version string
	// UserHome path of user home, same as ${HOME}
	UserHome string
	// AppHome path of kt config folder, default to ${UserHome}/.ktctl
	AppHome string
	// Component current sub-command (connect, exchange, mesh or provide)
	Component string
	// Shadow pod name
	Shadow string
	// Router pod name
	Router string
	// Mesh version of mesh pod
	Mesh string
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
	// SourceIP for tun mode
	SourceIP string
	// DestIP for tun mode
	DestIP   string
}

type dashboardOptions struct {
	Install bool
	Port    string
}

// DaemonOptions cli options
type DaemonOptions struct {
	RuntimeOptions    *RuntimeOptions
	ProvideOptions    *ProvideOptions
	ConnectOptions    *ConnectOptions
	ExchangeOptions   *ExchangeOptions
	MeshOptions       *MeshOptions
	CleanOptions      *CleanOptions
	DashboardOptions  *dashboardOptions
	KubeConfig        string
	Namespace         string
	ServiceAccount    string
	Debug             bool
	Image             string
	ImagePullSecret   string
	NodeSelector      string
	WithLabels        string
	WithAnnotations   string
	KubeOptions       cli.StringSlice
	WaitTime          int
	AlwaysUpdateShadow bool
	UseKubectl        bool
	KubeContext       string
}

// NewDaemonOptions return new cli default options
func NewDaemonOptions(version string) *DaemonOptions {
	return &DaemonOptions{
		Namespace:  common.DefaultNamespace,
		KubeConfig: util.KubeConfig(),
		WaitTime:   5,
		RuntimeOptions: &RuntimeOptions{
			Version:  version,
			UserHome: util.UserHome,
			AppHome:  util.KtHome,
		},
		ConnectOptions:   &ConnectOptions{},
		ExchangeOptions:  &ExchangeOptions{},
		MeshOptions:      &MeshOptions{},
		CleanOptions:     &CleanOptions{},
		DashboardOptions: &dashboardOptions{},
		ProvideOptions:   &ProvideOptions{},
	}
}
