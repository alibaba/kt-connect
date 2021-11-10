package command

import (
	"context"
	"errors"
	"github.com/alibaba/kt-connect/pkg/kt/command/clean"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"k8s.io/api/apps/v1"
)

// newMeshCommand return new mesh command
func newMeshCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "mesh",
		Usage: "mesh kubernetes deployment to local",
		Flags: []urfave.Flag{
			urfave.StringFlag{
				Name:        "expose",
				Usage:       "ports to expose separate by comma, in [port] or [local:remote] format, e.g. 7001,8080:80",
				Destination: &options.MeshOptions.Expose,
			},
			urfave.StringFlag{
				Name:        "version-label",
				Usage:       "specify the version of mesh service, e.g. '0.0.1'",
				Destination: &options.MeshOptions.Version,
			},
			urfave.StringFlag{
				Name:        "method",
				Value:       "manual",
				Usage:       "Mesh method 'manual' or 'auto'(coming soon)",
				Destination: &options.MeshOptions.Method,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			deploymentToMesh := c.Args().First()
			expose := options.MeshOptions.Expose

			if len(deploymentToMesh) == 0 {
				return errors.New("name of deployment to mesh is required")
			}

			if len(expose) == 0 {
				return errors.New("--expose is required")
			}

			return action.Mesh(deploymentToMesh, cli, options)
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch, err := setupProcess(cli, options, common.ComponentMesh)
	if err != nil {
		return err
	}

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	ctx := context.Background()
	app, err := kubernetes.Deployment(ctx, deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	meshVersion := getVersion(options)

	shadowPodName := app.GetObjectMeta().GetName() + "-kt-" + meshVersion
	labels := getMeshLabels(shadowPodName, meshVersion, app, options)

	err = createShadowAndInbound(ctx, shadowPodName, labels, options, kubernetes)
	if err != nil {
		return err
	}

	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf("    Mesh Version '%s' You can update Istio rule       ", meshVersion)
	log.Info().Msg("---------------------------------------------------------")

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted: %s", <-process.Interrupt())
		clean.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func createShadowAndInbound(ctx context.Context, shadowPodName string, labels map[string]string, options *options.DaemonOptions,
	kubernetes cluster.KubernetesInterface) error {

	envs := make(map[string]string)
	annotations := make(map[string]string)
	podIP, podName, sshConfigMapName, _, err := kubernetes.GetOrCreateShadow(ctx, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}
	// record context data
	options.RuntimeOptions.Shadow = shadowPodName
	options.RuntimeOptions.SSHCM = sshConfigMapName

	shadow := connect.Create(options)
	err = shadow.Inbound(options.MeshOptions.Expose, podName, podIP)

	if err != nil {
		return err
	}
	return nil
}

func getMeshLabels(workload string, meshVersion string, app *v1.Deployment, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentMesh,
		common.KTName:      workload,
		common.KTVersion:   meshVersion,
	}
	if app != nil {
		for k, v := range app.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	return labels
}

func getVersion(options *options.DaemonOptions) string {
	if len(options.MeshOptions.Version) != 0 {
		return options.MeshOptions.Version
	}
	return strings.ToLower(util.RandomString(5))
}
