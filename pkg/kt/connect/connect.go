package connect

import (
	"github.com/alibaba/kt-connect/pkg/kt/options"
)

// Connect VPN connect interface
type Connect struct {
	Kubeconfig string
	Namespace  string
	Image      string
	Method     string
	Swap       string
	Expose     string
	ProxyPort  int
	Port       int // Local SSH Port
	DisableDNS bool
	PodCIDR    string
	Debug      bool
	PidFile    string
	Options    *options.DaemonOptions
}

