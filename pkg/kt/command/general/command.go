package general

import (
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/spf13/cobra"
)

func SimpleSubCommand(name, usage string, action func(args []string) error, postHandler func(cmd *cobra.Command)) *cobra.Command {
	cmd := &cobra.Command{
		Use: name,
		Short: usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.HideGlobalFlags(cmd)
			return action(args)
		},
	}
	cmd.Long = cmd.Short
	if postHandler != nil {
		postHandler(cmd)
	}
	cmd.SetUsageTemplate(UsageTemplate(false))
	return cmd
}
