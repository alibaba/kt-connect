package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/mesh"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"os"
)

// NewMeshCommand return new mesh command
func NewMeshCommand(action ActionInterface, ch chan os.Signal) urfave.Command {
	return urfave.Command{
		Name:  "mesh",
		Usage: "redirect marked requests of specified kubernetes service to local",
		UsageText: "ktctl mesh <service-name> [command options]",
		Flags: general.MeshActionFlag(opt.Get()),
		Action: func(c *urfave.Context) error {
			if opt.Get().Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(); err != nil {
				return err
			}

			if len(c.Args()) == 0 {
				return fmt.Errorf("name of deployment to mesh is required")
			}
			if len(opt.Get().MeshOptions.Expose) == 0 {
				return fmt.Errorf("--expose is required")
			}

			return action.Mesh(c.Args().First(), ch)
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(resourceName string, ch chan os.Signal) error {
	err := general.SetupProcess(util.ComponentMesh, ch)
	if err != nil {
		return err
	}

	if port := util.FindBrokenLocalPort(opt.Get().MeshOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	// Get service to mesh
	svc, err := general.GetServiceByResourceName(resourceName, opt.Get().Namespace)
	if err != nil {
		return err
	}

	if port := util.FindInvalidRemotePort(opt.Get().MeshOptions.Expose, svc.Spec.Ports); port != "" {
		return fmt.Errorf("target port %s not exists in service %s", port, svc.Name)
	}

	if opt.Get().MeshOptions.Mode == util.MeshModeManual {
		err = mesh.ManualMesh(svc)
	} else if opt.Get().MeshOptions.Mode == util.MeshModeAuto {
		err = mesh.AutoMesh(svc)
	} else {
		err = fmt.Errorf("invalid mesh method '%s', supportted are %s, %s", opt.Get().MeshOptions.Mode,
			util.MeshModeAuto, util.MeshModeManual)
	}
	if err != nil {
		return err
	}

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

