package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/config"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/spf13/cobra"
)

// NewConfigCommand return new config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "config",
		Short: "List, get or set default value for command options",
		RunE: func(cmd *cobra.Command, args []string) error {
			hideGlobalFlags(cmd)
			return cmd.Help()
		},
	}

	cmd.AddCommand(SubConfig("show", "List all available and configured options", config.Show, config.ShowHandle))
	cmd.AddCommand(SubConfig("get", "Fetch default value of specified option", config.Get, nil))
	cmd.AddCommand(SubConfig("set", "Customize default value of specified option", config.Set, nil))
	cmd.AddCommand(SubConfig("reset", "Restore default value of specified option", config.Reset, config.ResetHandle))
	cmd.AddCommand(SubConfig("save", "Save current configured options as a profile", config.Save, config.SaveHandle))
	cmd.AddCommand(SubConfig("load", "Show profiles or load config from a profile", config.Load, nil))

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl config [list | get key | set key=value | reset key]"))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Config, opt.ConfigFlags())
	return cmd
}

func hideGlobalFlags(cmd *cobra.Command) {
	for _, f := range opt.GlobalFlags() {
		_ = cmd.InheritedFlags().MarkHidden(util.UnCapitalize(f.Target))
	}
}

func SubConfig(name, usage string, action func(args []string) error, postHandler func(cmd *cobra.Command)) *cobra.Command {
	cmd := &cobra.Command{
		Use: name,
		Short: usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			hideGlobalFlags(cmd)
			return action(args)
		},
	}
	cmd.Long = cmd.Short
	if postHandler != nil {
		postHandler(cmd)
	}
	return cmd
}
