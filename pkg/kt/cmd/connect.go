package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	connectExample = `
  # connect to kubernetes cluster
  kubectl connect
  # connect with debug mode
  kubectl connect -d
  # connect with socks
  kubectl connect -m socks --dump2hosts=default,dev
  # connect with socks and dump service to local hosts file
`
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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
	cmd.Flags().IntVarP(&opt.Timeout, "timeout", "", 30, "timeout to wait port-forward")

	// method
	cmd.Flags().StringVarP(&opt.Method, "method", "m", "", "connect provider vpn/socks/socks5/tun(alpha)")
	cmd.Flags().IntVarP(&opt.Port, "port", "p", 2222, "Local SSH Proxy port ")
	cmd.Flags().BoolVarP(&opt.Global, "global", "g", false, "with cluster scope")

	// vpn
	cmd.Flags().BoolVarP(&opt.DisableDNS, "disableDNS", "", false, "disable Cluster DNS")
	cmd.Flags().StringVarP(&opt.Cidr, "cidr", "c", "", "Custom CIDR, e.g. '172.2.0.0/16")

	// tun
	cmd.Flags().StringVarP(&opt.TunName, "tunName", "", "tun0", "The tun device name to create on client machine (Alpha). Only works on Linux")
	cmd.Flags().StringVarP(&opt.TunCidr, "tunCidr", "", "10.1.1.0/30", "The cidr used by local tun and peer tun device, at least 4 ips. This cidr MUST NOT overlay with kubernetes service cidr and pod cidr")

	// socks
	cmd.Flags().IntVarP(&opt.Proxy, "proxy", "", 2223, "when should method socks or socks5, you can choice which port to proxy")
	cmd.Flags().StringVarP(&opt.Dump2hosts, "dump2hosts", "", "", "specify namespaces to dump service into local hosts file")

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
	o.restConfig = restConfig
	return nil
}

// Run ...
func (o *ConnectOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}

	ops := o.transport()
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

// CloneDaemonOptions ...
func (o *ConnectOptions) transport() *options.DaemonOptions {
	daemonOptions := o.transportGlobalOptions()
	daemonOptions.ConnectOptions = &options.ConnectOptions{
		DisableDNS:           o.DisableDNS,
		Method:               o.Method,
		SocksPort:            o.Proxy,
		CIDR:                 o.Cidr,
		SSHPort:              o.Port,
		Dump2HostsNamespaces: strings.Split(o.Dump2hosts, ","),
		TunName:              o.TunName,
		TunCidr:              o.TunCidr,
	}
	return daemonOptions
}

// NewConnectOptions ...
func NewConnectOptions(streams genericclioptions.IOStreams) *ConnectOptions {
	return &ConnectOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}
