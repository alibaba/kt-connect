package options

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// ConnectOptions ...
type ConnectOptions struct {
	Global           bool
	DisablePodIp     bool
	DisableTunDevice bool
	DisableTunRoute  bool
	SocksPort        int
	DnsPort          int
	DnsCacheTtl      int
	IncludeIps       string
	ExcludeIps       string
	Mode             string
	DnsMode          string
	SharedShadow     bool
	ClusterDomain    string
	SkipCleanup      bool
}

// ExchangeOptions ...
type ExchangeOptions struct {
	Mode            string
	Expose          string
	RecoverWaitTime int
}

// MeshOptions ...
type MeshOptions struct {
	Mode        string
	Expose      string
	VersionMark string
	RouterImage string
}

// RecoverOptions ...
type RecoverOptions struct {

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
	SweepLocalRoute  bool
}

// ConfigOptions ...
type ConfigOptions struct {

}

// GlobalOptions ...
type GlobalOptions struct {
	RunAsWorkerProcess  bool
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
	UseShadowDeployment bool
	AlwaysUpdateShadow  bool
	SkipTimeDiff        bool
	KubeContext         string
	PodQuota            string
}

// RuntimeStore ...
type RuntimeStore struct {
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
	// Origin the origin deployment or service name
	Origin string
	// Replicas the origin replicas
	Replicas int32
	// Service exposed service name
	Service string
	// RestConfig kubectl config
	RestConfig *rest.Config
}

// DaemonOptions cli options
type DaemonOptions struct {
	Runtime  *RuntimeStore
	Preview  *PreviewOptions
	Connect  *ConnectOptions
	Exchange *ExchangeOptions
	Mesh     *MeshOptions
	Recover  *RecoverOptions
	Clean    *CleanOptions
	Config   *ConfigOptions
	Global   *GlobalOptions
}

var opt *DaemonOptions

// Get fetch options instance
func Get() *DaemonOptions {
	if opt == nil {
		opt = &DaemonOptions{
			Global: &GlobalOptions {
				Namespace:  util.DefaultNamespace,
			},
			Runtime: &RuntimeStore{
				UserHome: util.UserHome,
				AppHome:  util.KtHome,
			},
			Connect:  &ConnectOptions{},
			Exchange: &ExchangeOptions{},
			Mesh:     &MeshOptions{},
			Preview:  &PreviewOptions{},
			Clean:    &CleanOptions{},
		}
	}
	return opt
}
