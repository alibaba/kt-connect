package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/spf13/cobra"
	"strings"
)

// NewReceiveCommand return new Receive command
func NewReceiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "receive",
		Short: "Redirect a local port to specified kubernetes service, domain or ip address",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a service name or target address must be specified")
			} else if len(args) > 1 {
				return fmt.Errorf("too many target addresses are spcified (%s), should be one", strings.Join(args, ",") )
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Receive(args[0])
		},
		Example: "ktctl receive <domain|ip> [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Receive, opt.ReceiveFlags())
	return cmd
}

func Receive(host string) error {
	return nil
}
