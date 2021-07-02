package cmd

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/command"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	meshExample = `
  # mesh app to local
  kubectl mesh tomcat --expose 8080
`
)

// NewMeshCommand ...
func NewMeshCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	opt := NewMeshOptions(streams)

	cmd := &cobra.Command{
		Use:          "mesh",
		Short:        "mesh app",
		Example:      meshExample,
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
	cmd.Flags().StringVarP(&opt.Version, "version-label", "", "0.0.1", "specify the version of mesh service, e.g. '0.0.1'")
	return cmd
}

// MeshOptions ...
type MeshOptions struct {
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

	// mesh
	Target  string
	Expose  string
	Version string
}

// NewMeshOptions ...
func NewMeshOptions(streams genericclioptions.IOStreams) *MeshOptions {
	return &MeshOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// Complete ...
func (o *MeshOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	if len(o.args) < 2 {
		return fmt.Errorf("missing mesh target")
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
func (o *MeshOptions) Run() error {
	if err := o.checkContext(); err != nil {
		return err
	}
	if err := o.checkTarget(); err != nil {
		return err
	}

	ops := o.transport()
	context := &kt.Cli{Options: ops}
	action := command.Action{}

	return action.Mesh(o.Target, context, ops)
}

// checkContext
func (o *MeshOptions) checkContext() error {
	currentCtx := o.rawConfig.CurrentContext
	if _, ok := o.rawConfig.Contexts[currentCtx]; !ok {
		return fmt.Errorf("current context %s not found anymore in KUBECONFIG", currentCtx)
	}
	return nil
}

// checkTarget
func (o *MeshOptions) checkTarget() error {
	if _, err := o.clientset.AppsV1().Deployments(o.currentNs).Get(o.Target, metav1.GetOptions{}); err != nil {
		return err
	}
	return nil
}

func (o *MeshOptions) transport() *options.DaemonOptions {
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
		MeshOptions: &options.MeshOptions{
			Expose:  o.Expose,
			Version: o.Version,
		},
		ConnectOptions: &options.ConnectOptions{},
	}
}
