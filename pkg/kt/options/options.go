package options

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// RunOptions ...
type RunOptions struct {
	Expose bool
	Port   int
}

type connectOptions struct {
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
	Expose string
}

type runtimeOptions struct {
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
	Replicas int32
	Service  string
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
	RuntimeOptions   *runtimeOptions
	RunOptions       *RunOptions
	ConnectOptions   *connectOptions
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
		RuntimeOptions: &runtimeOptions{
			UserHome: userHome,
			AppHome:  appHome,
			PidFile:  pidFile,
		},
		ConnectOptions:   &connectOptions{},
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
