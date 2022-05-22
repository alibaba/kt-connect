package command

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/command/preview"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

// NewPreviewCommand return new preview command
func NewPreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Expose a local service to kubernetes cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("a service name must be specified")
			} else if len(args) > 1 {
				return fmt.Errorf("too many service name are spcified (%s), should be one", strings.Join(args, ",") )
			}
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Preview(args[0])
		},
		Example: "ktctl preview <service-name> [command options]",
	}

	cmd.SetUsageTemplate(general.UsageTemplate(true))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().Preview, opt.PreviewFlags())
	return cmd
}

// Preview create a new service in cluster
func Preview(serviceName string) error {
	ch, err := general.SetupProcess(util.ComponentPreview)
	if err != nil {
		return err
	}

	if opt.Get().Mesh.SkipPortChecking {
		if port := util.FindBrokenLocalPort(opt.Get().Preview.Expose); port != "" {
			return fmt.Errorf("no application is running on port %s", port)
		}
	}

	if err = preview.Expose(serviceName); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your local service in cluster by name '%s'", serviceName)
	log.Info().Msg("---------------------------------------------------------------")

	// watch background process, clean the workspace and exit if background process occur exception
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return nil
}
