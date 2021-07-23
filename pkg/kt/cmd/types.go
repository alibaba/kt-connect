package cmd

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// GlobalOptions ...
type GlobalOptions struct {
	// global
	Labels    string
	Image     string
	Debug     bool
	currentNs string
	Timeout   int

	// common
	args                   []string
	userSpecifiedNamespace string
	restConfig             *rest.Config
	rawConfig              api.Config
	clientset              kubernetes.Interface
}

// ExchangeOptions ...
type ExchangeOptions struct {
	GlobalOptions

	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	// exchange
	Target string
	Expose string
}

// ConnectOptions ...
type ConnectOptions struct {
	GlobalOptions

	genericclioptions.IOStreams
	configFlags *genericclioptions.ConfigFlags

	Method     string
	Labels     string
	Proxy      int
	DisableDNS bool
	Cidr       string
	Dump2hosts string
	Port       int
	Global     bool
}

// MeshOptions ...
type MeshOptions struct {
	GlobalOptions

	genericclioptions.IOStreams
	configFlags *genericclioptions.ConfigFlags

	// mesh
	Target  string
	Expose  string
	Version string
}

// ProvideOptions ...
type ProvideOptions struct {
	GlobalOptions

	genericclioptions.IOStreams
	configFlags *genericclioptions.ConfigFlags

	// run
	Expose   int
	External bool
	Target   string
}
