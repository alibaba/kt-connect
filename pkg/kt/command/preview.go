package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/preview"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
	"time"
)

// NewPreviewCommand return new preview command
func NewPreviewCommand(action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "preview",
		Usage: "expose a local service to kubernetes cluster",
		UsageText: "ktctl preview <service-name> [command options]",
		Flags: general.PreviewActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}
			if len(c.Args()) == 0 {
				return errors.New("an service name must be specified")
			}
			if len(opt.Get().PreviewOptions.Expose) == 0 {
				return errors.New("--expose is required")
			}
			return action.Preview(c.Args().First())
		},
	}
}

// Preview create a new service in cluster
func (action *Action) Preview(serviceName string) error {
	ch, err := general.SetupProcess(common.ComponentPreview)
	if err != nil {
		return err
	}

	if port := util.FindBrokenPort(opt.Get().PreviewOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	if err = preview.Expose(context.TODO(), serviceName); err != nil {
		return err
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace()
		os.Exit(0)
	}()

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	// when process interrupt by signal, wait a while for resource clean up
	time.Sleep(1 * time.Second)
	return nil
}
