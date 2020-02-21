package options

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

type connectOptions struct {
	DisableDNS  bool
	SSHPort     int
	Socke5Proxy int
	CIDR        string
	Method      string
	Dump2Hosts  bool
	Hosts       map[string]string
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
	// The origin app name
	Origin string
	// The origin repicas
	Replicas int32
}

type dashboardOptions struct {
	Install bool
}

// DaemonOptions cli options
type DaemonOptions struct {
	KubeConfig       string
	Namespace        string
	Debug            bool
	Image            string
	Labels           string
	RuntimeOptions   *runtimeOptions
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
	}
}
