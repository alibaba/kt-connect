package options

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/spf13/cobra"
)

func HideGlobalFlags(cmd *cobra.Command) {
	for _, f := range GlobalFlags() {
		_ = cmd.InheritedFlags().MarkHidden(util.UnCapitalize(f.Target))
	}
}
