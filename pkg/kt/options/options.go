package options

import (
	"fmt"
	"path/filepath"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/kt/vars"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
)

// RunOptions ...
type RunOptions struct {
	Expose bool
	Port   int
}

// ConnectOptions ...
type ConnectOptions struct {
	DisableDNS           bool
	SSHPort              int
	Socke5Proxy          int
	CIDR                 string
	Method               string
	Dump2Hosts           bool
	Dump2HostsNamespaces cli.StringSlice
	Hosts                map[string]string
	ShareShadow          bool
}

type exchangeOptions struct {
	Expose string
}

type meshOptions struct {
	Expose  string
	Version string
}

// RuntimeOptions ...
type RuntimeOptions struct {
	PidFile  string
	UserHome string
	AppHome  string
	// Shadow deployment name
	Shadow string
	// ssh public key name of config map. format is kt-xxx(component)-public-key-xxx(version)
	SSHCM string
	// The origin app name
	Origin string
	// The origin repicas
	Replicas  int32
	Service   string
	Clientset kubernetes.Interface
}

type dashboardOptions struct {
	Install bool
	Port    string
}

// DaemonOptions cli options
type DaemonOptions struct {
	KubeConfig       string
	Namespace        string
	Debug            bool
	Image            string
	Labels           string
	KubeOptions      cli.StringSlice
	RuntimeOptions   *RuntimeOptions
	RunOptions       *RunOptions
	ConnectOptions   *ConnectOptions
	ExchangeOptions  *exchangeOptions
	MeshOptions      *meshOptions
	DashboardOptions *dashboardOptions
}

// NewDaemonOptions return new cli default options
func NewDaemonOptions() *DaemonOptions {
	userHome := util.HomeDir()
	appHome := fmt.Sprintf("%s/.ktctl", userHome)
	util.CreateDirIfNotExist(appHome)
	pidFile := fmt.Sprintf("%s/pid", appHome)
	return &DaemonOptions{
		Namespace:  vars.DefNamespace,
		KubeConfig: filepath.Join(userHome, ".kube", "config"),
		RuntimeOptions: &RuntimeOptions{
			UserHome: userHome,
			AppHome:  appHome,
			PidFile:  pidFile,
		},
		ConnectOptions:   &ConnectOptions{},
		ExchangeOptions:  &exchangeOptions{},
		MeshOptions:      &meshOptions{},
		DashboardOptions: &dashboardOptions{},
		RunOptions:       &RunOptions{},
	}
}

// NewRunDaemonOptions ...
func NewRunDaemonOptions(labels string, options *RunOptions) *DaemonOptions {
	daemonOptions := NewDaemonOptions()
	daemonOptions.Labels = labels
	daemonOptions.RunOptions = options
	return daemonOptions
}
