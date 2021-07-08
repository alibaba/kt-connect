package command

import (
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"

	"github.com/alibaba/kt-connect/pkg/kt"

	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// newRunCommand return new run command
func newRunCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "run",
		Usage: "create a shadow deployment to redirect request to user local",
		Flags: []urfave.Flag{
			urfave.IntFlag{
				Name:        "port",
				Usage:       "The port that exposes",
				Destination: &options.RunOptions.Port,
			},
			urfave.BoolFlag{
				Name:        "expose",
				Usage:       "If true, a public, external service is created",
				Destination: &options.RunOptions.Expose,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			port := options.RunOptions.Port
			if len(c.Args()) == 0 {
				return errors.New("an identifier name must be provided")
			}
			if port == 0 {
				return errors.New("--port is required")
			}
			return action.Run(c.Args().First(), cli, options)
		},
	}
}

// Run create a new service in cluster
func (action *Action) Run(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch := SetUpCloseHandler(cli, options, "run")
	if err := run(deploymentName, cli, options); err != nil {
		return err
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-util.Interrupt()
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	<-ch
	return nil
}

// Run create a new service in cluster
func run(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	version := strings.ToLower(util.RandomString(5))
	name := fmt.Sprintf("%s-kt-%s", deploymentName, version)
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentRun,
		common.KTName:      name,
		common.KTVersion:   version,
	}
	annotations := map[string]string{
		common.KTConfig: fmt.Sprintf("expose=%t", options.RunOptions.Expose),
	}

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	return runAndExposeLocalService(name, labels, annotations, options, kubernetes, cli)
}

// runAndExposeLocalService create shadow and expose service if need
func runAndExposeLocalService(name string, labels, annotations map[string]string, options *options.DaemonOptions,
	kubernetes cluster.KubernetesInterface, cli kt.CliInterface) (err error) {

	envs := make(map[string]string)
	podIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(
		name, options.Namespace, options.Image, labels, annotations, envs, options.Debug, false)
	if err != nil {
		return err
	}
	log.Info().Msgf("create shadow pod %s ip %s", podName, podIP)

	if options.RunOptions.Expose {
		log.Info().Msgf("expose deployment %s to service %s:%v", name, name, options.RunOptions.Port)
		_, err = kubernetes.CreateService(name, options.Namespace, options.RunOptions.Port, labels)
		if err != nil {
			return err
		}
		options.RuntimeOptions.Service = name
	}

	options.RuntimeOptions.Shadow = name
	options.RuntimeOptions.SSHCM = sshcm

	err = cli.Shadow().Inbound(strconv.Itoa(options.RunOptions.Port), podName, podIP, credential)
	if err != nil {
		return err
	}

	log.Info().Msgf("forward remote %s:%v -> 127.0.0.1:%v", podIP, options.RunOptions.Port, options.RunOptions.Port)
	return nil
}
