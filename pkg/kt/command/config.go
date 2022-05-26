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
		Example: "ktctl config <sub-command> [options]",
	}

	cmd.AddCommand(general.SimpleSubCommand("show", "Show all available and configured options", config.Show, config.ShowHandle))
	cmd.AddCommand(general.SimpleSubCommand("get", "Fetch default value of specified option", config.Get, nil))
	cmd.AddCommand(general.SimpleSubCommand("set", "Customize default value of specified option", config.Set, nil))
	cmd.AddCommand(general.SimpleSubCommand("unset", "Restore default value of specified option", config.Unset, config.UnsetHandle))
	cmd.AddCommand(general.SimpleSubCommand("list-profile", "List all pre-saved profiles", config.ListProfile, nil))
	cmd.AddCommand(general.SimpleSubCommand("save-profile", "Save current configured options as a profile", config.SaveProfile, nil))
	cmd.AddCommand(general.SimpleSubCommand("load-profile", "Load config from a profile", config.LoadProfile, config.LoadProfileHandle))
	cmd.AddCommand(general.SimpleSubCommand("drop-profile", "Delete a profile", config.DropProfile, nil))

	cmd.SetUsageTemplate(general.UsageTemplate(false))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Config, opt.ConfigFlags())
	return cmd
}
