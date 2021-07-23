package cmd

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	exchangeExample = `
  # exchange app to local
  kubectl exchange tomcat --expose 8080
`
)

// NewExchangeCommand ...
func NewExchangeCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	opt := NewExchangeOptions(streams)

	cmd := &cobra.Command{
		Use:          "exchange",
		Short:        "exchange app",
		Example:      exchangeExample,
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

	// exchange
	cmd.Flags().StringVarP(&opt.Expose, "expose", "", "80", " expose port [port] or [remote:local]")

	return cmd
}

// NewExchangeOptions ...
func NewExchangeOptions(streams genericclioptions.IOStreams) *ExchangeOptions {
	return &ExchangeOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// Complete ...
func (o *ExchangeOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	if len(o.args) < 2 {
		return fmt.Errorf("missing exchange target")
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
func (o *ExchangeOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}
	if err := o.checkTarget(); err != nil {
		return err
	}

	ops := o.transport()
	context := &kt.Cli{Options: ops}
	action := command.Action{}

	return action.Exchange(o.Target, context, ops)
}

// checkContext
func (o *ExchangeOptions) checkContext() error {
	currentCtx := o.rawConfig.CurrentContext
	if _, ok := o.rawConfig.Contexts[currentCtx]; !ok {
		return fmt.Errorf("current context %s not found anymore in KUBECONFIG", currentCtx)
	}
	return nil
}

// checkTarget
func (o *ExchangeOptions) checkTarget() error {
	if _, err := o.clientset.AppsV1().Deployments(o.currentNs).Get(o.Target, metav1.GetOptions{}); err != nil {
		return err
	}
	return nil
}

func (o *ExchangeOptions) transport() *options.DaemonOptions {
	return &options.DaemonOptions{
		Image:     o.Image,
		Debug:     o.Debug,
		Labels:    o.Labels,
		Namespace: o.currentNs,
		WaitTime:  o.Timeout,
		RuntimeOptions: &options.RuntimeOptions{
			UserHome:   userHome,
			AppHome:    appHome,
			PidFile:    pidFile,
			Clientset:  o.clientset,
			RestConfig: o.restConfig,
		},
		ExchangeOptions: &options.ExchangeOptions{
			Expose: o.Expose,
		},
		ConnectOptions: &options.ConnectOptions{},
	}
}
