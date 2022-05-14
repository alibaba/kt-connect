package command

import (
	"fmt"
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
			for _, f := range opt.GlobalFlags() {
				_ = cmd.InheritedFlags().MarkHidden(f.Name)
			}
			return cmd.Help()
		},
	}

	cmd.AddCommand(SubConfig("list", "Show all available options", config.List))
	cmd.AddCommand(SubConfig("get", "Fetch default value of specified option", config.Get))
	cmd.AddCommand(SubConfig("set", "Customize default value of specified option", config.Set))
	cmd.AddCommand(SubConfig("reset", "Restore default value of specified option", config.Reset))

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl config [list | get key | set key=value | reset key]"))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Config, opt.ConfigFlags())
	return cmd
}

func SubConfig(name, usage string, action func(args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use: name,
		Short: usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action(args)
		},
	}
	cmd.Long = cmd.Short
	return cmd
}
