package cmd

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	provideExample = `
  # expose local service to cluster 
  kubectl provide tomcat --expose 80
`
)

// NewProvideCommand ...
func NewProvideCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	opt := NewProvideOptions(streams)

	cmd := &cobra.Command{
		Use:          "provide",
		Short:        "provide app",
		Example:      provideExample,
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

	// run
	cmd.Flags().IntVarP(&opt.Expose, "expose", "", 80, " The port that exposes")
	cmd.Flags().BoolVarP(&opt.External, "external", "e", false, " If specified, a public, external service is created")
	return cmd
}

// NewProvideOptions ...
func NewProvideOptions(streams genericclioptions.IOStreams) *ProvideOptions {
	return &ProvideOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// Complete ...
func (o *ProvideOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	if len(o.args) < 2 {
		return fmt.Errorf("please specify service name")
	}

	o.Target = args[1]

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
	if o.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return nil
}

// Run ...
func (o *ProvideOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}
	ops := o.transport()
	context := &kt.Cli{Options: ops}
	action := command.Action{}

	return action.Provide(o.Target, context, ops)
}

// checkContext
func (o *ProvideOptions) checkContext() error {
	currentCtx := o.rawConfig.CurrentContext
	if _, ok := o.rawConfig.Contexts[currentCtx]; !ok {
		return fmt.Errorf("current context %s not found anymore in KUBECONFIG", currentCtx)
	}
	return nil
}

func (o *ProvideOptions) transport() *options.DaemonOptions {
	daemonOptions := o.transportGlobalOptions()
	daemonOptions.ProvideOptions = &options.ProvideOptions{
		External: o.External,
		Expose:   o.Expose,
	}
	return daemonOptions
}
