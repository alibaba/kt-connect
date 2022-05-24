package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/command/config"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/spf13/cobra"
)

// NewConfigCommand return new config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "config",
		Short: "List, get or set default value for command options",
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.HideGlobalFlags(cmd)
			return cmd.Help()
		},
		Example: "ktctl config [show | get <key> | set <key> <value> | unset <key>]",
	}

	cmd.AddCommand(general.SimpleSubCommand("show", "Show all available and configured options", config.Show, config.ShowHandle))
	cmd.AddCommand(general.SimpleSubCommand("get", "Fetch default value of specified option", config.Get, nil))
	cmd.AddCommand(general.SimpleSubCommand("set", "Customize default value of specified option", config.Set, nil))
	cmd.AddCommand(general.SimpleSubCommand("unset", "Restore default value of specified option", config.Unset, config.UnsetHandle))

	cmd.SetUsageTemplate(general.UsageTemplate(false))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Config, opt.ConfigFlags())
	return cmd
}
