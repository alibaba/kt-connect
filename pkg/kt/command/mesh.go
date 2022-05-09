package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/mesh"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewMeshCommand return new mesh command
func NewMeshCommand(action ActionInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "mesh",
		Short: "Redirect marked requests of specified kubernetes service to local",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return general.Prepare()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("name of service to mesh is required")
			}
			return action.Mesh(args[0])
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl mesh <service-name> [command options]"))
	opt.SetOptions(cmd, cmd.Flags(), opt.Get().MeshOptions, opt.MeshFlags())
	return cmd
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(resourceName string) error {
	ch, err := general.SetupProcess(util.ComponentMesh)
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

	if port := util.FindInvalidRemotePort(opt.Get().MeshOptions.Expose, general.GetTargetPorts(svc)); port != "" {
		return fmt.Errorf("target port %s not exists in service %s", port, svc.Name)
	}

	log.Info().Msgf("Using %s mode", opt.Get().MeshOptions.Mode)
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
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return nil
}

