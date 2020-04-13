package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	connectExample = `
  # connect to kubernetes cluster
  kubectl connect
  # connect with debug mode
  kubectl connect -d
  # connect with socks5
  kubectl connect -m socks5 --dump2hosts=default,dev
  # connect with socks5 and dump service to local hosts file
`
)

// ConnectOptions ...
type ConnectOptions struct {
	configFlags *genericclioptions.ConfigFlags
	rawConfig   api.Config
	args        []string

	userSpecifiedNamespace string
	genericclioptions.IOStreams
	clientset  *kubernetes.Clientset
	image      string
	method     string
	debug      bool
	labels     string
	proxy      int
	disableDNS bool
	cidr       string
	dump2hosts string
}

// NewConnectCommand ...
func NewConnectCommand(streams genericclioptions.IOStreams) *cobra.Command {
	opt := NewConnectOptions(streams)

	cmd := &cobra.Command{
		Use:          "connect",
		Short:        "connect to cluster",
		Example:      connectExample,
		SilenceUsage: true,
		Version:      "0.0.1",
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
	cmd.Flags().BoolVarP(&opt.debug, "debug", "d", false, "debug mode")
	cmd.Flags().StringVarP(&opt.image, "image", "i", "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-shadow", "shadow image")
	cmd.Flags().StringVarP(&opt.labels, "labels", "l", "", "custom labels on shadow pod")

	// connect options
	cmd.Flags().StringVarP(&opt.method, "method", "m", "", "connect provider vpn/socks5")
	cmd.Flags().IntVarP(&opt.proxy, "proxy", "p", 2222, "when should method socks5, you can choice which port to proxy")
	cmd.Flags().BoolVarP(&opt.disableDNS, "disableDNS", "", false, "disable Cluster DNS")
	cmd.Flags().StringVarP(&opt.cidr, "cidr", "c", "", "Custom CIDR eq '172.2.0.0/16")
	cmd.Flags().StringVarP(&opt.dump2hosts, "dump2hosts", "", "", "auto write service to local hosts file")

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

	currentNS := o.rawConfig.Contexts[o.rawConfig.CurrentContext].Namespace
	fmt.Println(currentNS)
	return nil
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
