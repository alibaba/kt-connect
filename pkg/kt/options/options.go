package options

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// ConnectOptions ...
type ConnectOptions struct {
	Global               bool
	DisableDNS           bool
	DisablePodIp         bool
	SSHPort              int
	SocksPort            int
	IncludeIps           string
	ExcludeIps           string
	Mode                 string
	Dump2HostsNamespaces string
	SoleShadow           bool
	ClusterDomain        string
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Mode   string
	Expose string
	RecoverWaitTime int
}

// MeshOptions ...
type MeshOptions struct {
	Mode   string
	Expose string
	VersionMark string
	RouterImage string
}

// PreviewOptions ...
type PreviewOptions struct {
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
	// Component current sub-command (connect, exchange, mesh or preview)
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
	// RestConfig kubectl config
	RestConfig *rest.Config
}

// DaemonOptions cli options
type DaemonOptions struct {
	RuntimeOptions      *RuntimeOptions
	PreviewOptions      *PreviewOptions
	ConnectOptions      *ConnectOptions
	ExchangeOptions     *ExchangeOptions
	MeshOptions         *MeshOptions
	CleanOptions        *CleanOptions
	KubeConfig          string
	Namespace           string
	ServiceAccount      string
	Debug               bool
	Image               string
	ImagePullSecret     string
	NodeSelector        string
	WithLabels          string
	WithAnnotations     string
	PortForwardWaitTime int
	PodCreationWaitTime int
	AlwaysUpdateShadow  bool
	KubeContext         string
}

// NewDaemonOptions return new cli default options
func NewDaemonOptions(version string) *DaemonOptions {
	return &DaemonOptions{
		Namespace:  common.DefaultNamespace,
		KubeConfig: util.KubeConfig(),
		RuntimeOptions: &RuntimeOptions{
			Version:  version,
			UserHome: util.UserHome,
			AppHome:  util.KtHome,
		},
		ConnectOptions:  &ConnectOptions{},
		ExchangeOptions: &ExchangeOptions{},
		MeshOptions:     &MeshOptions{},
		PreviewOptions:  &PreviewOptions{},
		CleanOptions:    &CleanOptions{},
	}
}
