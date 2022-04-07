package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/preview"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

// NewPreviewCommand return new preview command
func NewPreviewCommand(action ActionInterface, ch chan os.Signal) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "preview",
		Long: "Expose a local service to kubernetes cluster",
		Short: "ktctl preview <service-name> [command options]",
		Run: func(cmd *cobra.Command, args []string) {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				log.Error().Msgf("%s", err)
			} else if len(args) == 0 {
				log.Error().Msgf("an service name must be specified")
			} else if err2 := action.Preview(args[0], ch); err2 != nil {
				log.Error().Msgf("%s", err2)
			}
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&opt.Get().PreviewOptions.Expose, "expose", "", "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80")
	cmd.Flags().BoolVar(&opt.Get().PreviewOptions.External, "external", false, "If specified, a public, external service is created")
	_ = cmd.MarkFlagRequired("expose")
	return cmd
}

// Preview create a new service in cluster
func (action *Action) Preview(serviceName string, ch chan os.Signal) error {
	err := general.SetupProcess(util.ComponentPreview, ch)
	if err != nil {
		return err
	}

	if port := util.FindBrokenLocalPort(opt.Get().PreviewOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	if err = preview.Expose(serviceName); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your local service in cluster by name '%s'", serviceName)
	log.Info().Msg("---------------------------------------------------------------")

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		ch <-os.Interrupt
	}()

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return nil
}
