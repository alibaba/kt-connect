package command

import (
	"errors"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	v1 "k8s.io/api/apps/v1"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// ComponentMesh mesh component
const ComponentMesh = "mesh"

// KubernetesTool kt sign
const KubernetesTool = "kt"

// newMeshCommand return new mesh command
func newMeshCommand(options *options.DaemonOptions, action ActionInterface) cli.Command {
	return cli.Command{
		Name:  "mesh",
		Usage: "mesh kubernetes deployment to local",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "expose",
				Usage:       "expose port [port] or [remote:local]",
				Destination: &options.MeshOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			mesh := c.Args().First()
			expose := options.MeshOptions.Expose

			if len(mesh) == 0 {
				return errors.New("mesh target is required")
			}

			if len(expose) == 0 {
				return errors.New("-expose is required")
			}
			return action.Mesh(mesh, options)
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(mesh string, options *options.DaemonOptions) error {
	checkConnectRunning(options.RuntimeOptions.PidFile)

	ch := SetUpCloseHandler(options)

	kubernetes, err := cluster.Create(options.KubeConfig)
	if err != nil {
		return err
	}

	app, err := kubernetes.Deployment(mesh, options.Namespace)
	if err != nil {
		return err
	}

	meshVersion := strings.ToLower(util.RandomString(5))
	workload := app.GetObjectMeta().GetName() + "-kt-" + meshVersion

	labels := getMeshLabels(workload, meshVersion, app, options)
	podIP, podName, sshcm, credential, err := kubernetes.CreateShadow(workload, options.Namespace, options.Image, labels, options.Debug)
	if err != nil {
		return err
	}

	// record context data
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	shadow := connect.Create(options)
	err = shadow.Inbound(options.MeshOptions.Expose, podName, podIP, credential)

	if err != nil {
		return err
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func getMeshLabels(workload string, meshVersion string, app *v1.Deployment, options *options.DaemonOptions) map[string]string {
	labels := map[string]string{
		"kt":           workload,
		"version":      meshVersion,
		"kt-component": ComponentMesh,
		"control-by":   KubernetesTool,
	}
	if app != nil {
		for k, v := range app.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}
	return labels
}
