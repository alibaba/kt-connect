package command

import (
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/command/profile"
	"github.com/spf13/cobra"
)

// NewProfileCommand return new profile command
func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "profile",
		Short: "Save config options as profile and load on demand",
		RunE: func(cmd *cobra.Command, args []string) error {
			opt.HideGlobalFlags(cmd)
			return cmd.Help()
		},
		Example: "ktctl profile [list | save <name> | load <name> | remove <name>]",
	}

	cmd.AddCommand(general.SimpleSubCommand("list", "List all pre-saved profiles", profile.List, nil))
	cmd.AddCommand(general.SimpleSubCommand("save", "Save current configured options as a profile", profile.Save, profile.SaveHandle))
	cmd.AddCommand(general.SimpleSubCommand("load", "Load config from a profile", profile.Load, profile.LoadHandle))
	cmd.AddCommand(general.SimpleSubCommand("remove", "Delete a profile", profile.Remove, nil))

	cmd.SetUsageTemplate(general.UsageTemplate(false))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Config, opt.ConfigFlags())
	return cmd
}
