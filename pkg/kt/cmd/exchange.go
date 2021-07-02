package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd/api"
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

	return cmd
}

// ExchangeOptions ...
type ExchangeOptions struct {
	configFlags *genericclioptions.ConfigFlags
	rawConfig   api.Config
	args        []string

	userSpecifiedNamespace string
	genericclioptions.IOStreams
	clientset kubernetes.Interface

	Timeout int
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
	return nil
}

// Run ...
func (o *ExchangeOptions) Run() error {
	return nil
}
