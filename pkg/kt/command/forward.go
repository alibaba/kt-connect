package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/spf13/cobra"
	"strings"
)

// NewForwardCommand return new Forward command
func NewForwardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward",
		Short: "Redirect local port to a service or any remote address",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a service name or target address must be specified")
			} else if len(args) > 1 {
				return fmt.Errorf("too many target addresses are spcified (%s), should be one", strings.Join(args, ",") )
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Forward(args[0])
		},
		Example: "ktctl forward <service-name|address:port> [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Forward, opt.ForwardFlags())
	return cmd
}

func Forward(target string) error {
	return nil
}
