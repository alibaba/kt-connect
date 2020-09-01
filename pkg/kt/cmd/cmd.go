package cmd

import (
	"fmt"
	"os"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	connectExample = `
  # connect to kubernetes cluster
  ktctl connect
  # connect with debug mode
  ktctl connect -d
  # connect with socks5
  ktctl connect -m socks5 --dump2hosts=default,dev
  # connect with socks5 and dump service to local hosts file
`
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// ConnectOptions ...
type ConnectOptions struct {
	configFlags *genericclioptions.ConfigFlags
	rawConfig   api.Config
	args        []string

	userSpecifiedNamespace string
	genericclioptions.IOStreams
	clientset  kubernetes.Interface
	Image      string
	Method     string
	Debug      bool
	Labels     string
	Proxy      int
	DisableDNS bool
	Cidr       string
	Dump2hosts string
	currentNs  string
	Port       int
}

// NewConnectCommand ...
func NewConnectCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	opt := NewConnectOptions(streams)

	cmd := &cobra.Command{
		Use:          "connect",
		Short:        "connect to cluster",
		Example:      connectExample,
		SilenceUsage: true,
		Version:      version,
		RunE: func(c *cobra.Command, args []string) error {
			if err := opt.Complete(c, args); err != nil {
				return err
			}
			if err := opt.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	// globals options
	cmd.Flags().StringVarP(&opt.currentNs, "namespace", "n", "", "current namespace")
	cmd.Flags().BoolVarP(&opt.Debug, "debug", "d", false, "debug mode")
	cmd.Flags().StringVarP(&opt.Image, "image", "i", "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow", "shadow image")
	cmd.Flags().StringVarP(&opt.Labels, "labels", "l", "", "custom labels on shadow pod")

	// method
	cmd.Flags().StringVarP(&opt.Method, "method", "m", "", "connect provider vpn/socks5")
	cmd.Flags().IntVarP(&opt.Port, "port", "p", 2222, "Local SSH Proxy port ")

	// vpn
	cmd.Flags().BoolVarP(&opt.DisableDNS, "disableDNS", "", false, "disable Cluster DNS")
	cmd.Flags().StringVarP(&opt.Cidr, "cidr", "c", "", "Custom CIDR, e.g. '172.2.0.0/16")

	// socks5
	cmd.Flags().IntVarP(&opt.Proxy, "proxy", "", 2223, "when should method socks5, you can choice which port to proxy")
	cmd.Flags().StringVarP(&opt.Dump2hosts, "dump2hosts", "", "", "auto write service to local hosts file")

	return cmd

}

// Complete ...
func (o *ConnectOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error
	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	if o.currentNs == "" {
		currentNS := o.rawConfig.Contexts[o.rawConfig.CurrentContext].Namespace
		if currentNS == "" {
			currentNS = "default"
		}
		o.currentNs = currentNS
	}

	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	o.clientset = clientset
	return nil
}

// Run ...
func (o *ConnectOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}

	ops := CloneDaemonOptions(o)
	context := &kt.Cli{Options: ops}
	action := command.Action{}

	if ops.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return action.Connect(context, ops)
}

func (o *ConnectOptions) checkContext() error {
	currentCtx := o.rawConfig.CurrentContext
	if _, ok := o.rawConfig.Contexts[currentCtx]; !ok {
		return fmt.Errorf("current context %s not found anymore in KUBECONFIG", currentCtx)
	}
	return nil
}

// NewConnectOptions ...
func NewConnectOptions(streams genericclioptions.IOStreams) *ConnectOptions {
	return &ConnectOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// CloneDaemonOptions ...
func CloneDaemonOptions(o *ConnectOptions) *options.DaemonOptions {
	userHome := util.HomeDir()
	appHome := fmt.Sprintf("%s/.ktctl", userHome)
	util.CreateDirIfNotExist(appHome)
	pidFile := fmt.Sprintf("%s/pid", appHome)
	return &options.DaemonOptions{
		Image:     o.Image,
		Debug:     o.Debug,
		Labels:    o.Labels,
		Namespace: o.currentNs,
		RuntimeOptions: &options.RuntimeOptions{
			UserHome:  userHome,
			AppHome:   appHome,
			PidFile:   pidFile,
			Clientset: o.clientset,
		},
		ConnectOptions: &options.ConnectOptions{
			DisableDNS:  o.DisableDNS,
			Method:      o.Method,
			Socke5Proxy: o.Proxy,
			CIDR:        o.Cidr,
			SSHPort:     o.Port,
		},
	}
}
