package command

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/command/mesh"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

// NewMeshCommand return new mesh command
func NewMeshCommand(action ActionInterface, ch chan os.Signal) *cobra.Command {
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
			return action.Mesh(args[0], ch)
		},
	}

	cmd.SetUsageTemplate(fmt.Sprintf(general.UsageTemplate, "ktctl mesh <service-name> [command options]"))
	cmd.Long = cmd.Short

	cmd.Flags().SortFlags = false
	cmd.InheritedFlags().SortFlags = false
	cmd.Flags().StringVar(&opt.Get().MeshOptions.Expose, "expose", "", "Ports to expose, use ',' separated, in [port] or [local:remote] format, e.g. 7001,8080:80")
	cmd.Flags().StringVar(&opt.Get().MeshOptions.Mode, "mode", util.MeshModeAuto, "Mesh method 'auto' or 'manual'")
	cmd.Flags().StringVar(&opt.Get().MeshOptions.VersionMark, "versionMark", "", "Specify the version of mesh service, e.g. '0.0.1' or 'mark:local'")
	cmd.Flags().StringVar(&opt.Get().MeshOptions.RouterImage, "routerImage", "registry.cn-hangzhou.aliyuncs.com/rdc-incubator/kt-connect-router:v" + opt.Get().RuntimeStore.Version, "(auto method only) Customize router image")
	_ = cmd.MarkFlagRequired("expose")
	return cmd
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

	if port := util.FindInvalidRemotePort(opt.Get().MeshOptions.Expose, general.GetTargetPorts(svc)); port != "" {
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

