package cmd

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	runExample = `
  # expose local service to cluster 
  kubectl run tomcat -e -p 80
`
)

// NewRunCommand ...
func NewRunCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	opt := NewRunOptions(streams)

	cmd := &cobra.Command{
		Use:          "mesh",
		Short:        "mesh app",
		Example:      runExample,
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
	cmd.Flags().IntVarP(&opt.Port, "port", "p", 80, " The port that exposes")
	cmd.Flags().BoolVarP(&opt.Expose, "expose", "e", false, " If true, a public, external service is created")
	return cmd
}

// RunOptions ...
type RunOptions struct {
	configFlags *genericclioptions.ConfigFlags
	rawConfig   api.Config
	args        []string

	userSpecifiedNamespace string
	genericclioptions.IOStreams
	clientset kubernetes.Interface

	// global
	Labels    string
	Image     string
	Debug     bool
	currentNs string
	Timeout   int

	// run
	Port   int
	Expose bool
	Target string
}

// NewRunOptions ...
func NewRunOptions(streams genericclioptions.IOStreams) *RunOptions {
	return &RunOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// Complete ...
func (o *RunOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	if len(o.args) < 2 {
		return fmt.Errorf("missing run target")
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
	if o.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	return nil
}

// Run ...
func (o *RunOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}
	ops := o.transport()
	context := &kt.Cli{Options: ops}
	action := command.Action{}

	return action.Run(o.Target, context, ops)
}

// checkContext
func (o *RunOptions) checkContext() error {
	currentCtx := o.rawConfig.CurrentContext
	if _, ok := o.rawConfig.Contexts[currentCtx]; !ok {
		return fmt.Errorf("current context %s not found anymore in KUBECONFIG", currentCtx)
	}
	return nil
}

func (o *RunOptions) transport() *options.DaemonOptions {
	userHome := util.HomeDir()
	appHome := fmt.Sprintf("%s/.ktctl", userHome)
	util.CreateDirIfNotExist(appHome)
	pidFile := fmt.Sprintf("%s/pid", appHome)
	return &options.DaemonOptions{
		Image:     o.Image,
		Debug:     o.Debug,
		Labels:    o.Labels,
		Namespace: o.currentNs,
		WaitTime:  o.Timeout,
		RuntimeOptions: &options.RuntimeOptions{
			UserHome:  userHome,
			AppHome:   appHome,
			PidFile:   pidFile,
			Clientset: o.clientset,
		},
		RunOptions: &options.RunOptions{
			Expose: o.Expose,
			Port:   o.Port,
		},
		ConnectOptions: &options.ConnectOptions{},
	}
}
