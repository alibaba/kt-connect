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

// newProvideCommand return new provide command
func newProvideCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "provide",
		Usage: "create a shadow service to redirect request to user local",
		Flags: []urfave.Flag{
			urfave.IntFlag{
				Name:        "expose",
				Usage:       "The port that exposes",
				Destination: &options.ProvideOptions.Expose,
			},
			urfave.BoolFlag{
				Name:        "external",
				Usage:       "If specified, a public, external service is created",
				Destination: &options.ProvideOptions.External,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			port := options.ProvideOptions.Expose
			if len(c.Args()) == 0 {
				return errors.New("an service name must be specified")
			}
			if port == 0 {
				return errors.New("--expose is required")
			}
			return action.Provide(c.Args().First(), cli, options)
		},
	}
}

// Provide create a new service in cluster
func (action *Action) Provide(serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch := SetUpCloseHandler(cli, options, "provide")
	if err := provide(serviceName, cli, options); err != nil {
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

// Provide create a new service in cluster
func provide(serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	version := strings.ToLower(util.RandomString(5))
	deploymentName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentProvide,
		common.KTName:      deploymentName,
		common.KTVersion:   version,
	}
	annotations := map[string]string{
		common.KTConfig: fmt.Sprintf("service=%s", serviceName),
	}

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	return exposeLocalService(serviceName, deploymentName, labels, annotations, options, kubernetes, cli)
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(serviceName, deploymentName string, labels, annotations map[string]string,
	options *options.DaemonOptions, kubernetes cluster.KubernetesInterface, cli kt.CliInterface) (err error) {

	envs := make(map[string]string)
	podIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(deploymentName, options, labels, annotations, envs, false)
	if err != nil {
		return err
	}
	log.Info().Msgf("create shadow pod %s ip %s", podName, podIP)

	log.Info().Msgf("expose deployment %s to service %s:%v", deploymentName, serviceName, options.ProvideOptions.Expose)
	_, err = kubernetes.CreateService(serviceName, options.Namespace, options.ProvideOptions.External, options.ProvideOptions.Expose, labels)
	if err != nil {
		return err
	}
	options.RuntimeOptions.Service = serviceName

	options.RuntimeOptions.Shadow = deploymentName
	options.RuntimeOptions.SSHCM = sshcm

	err = cli.Shadow().Inbound(strconv.Itoa(options.ProvideOptions.Expose), podName, podIP, credential)
	if err != nil {
		return err
	}

	log.Info().Msgf("forward remote %s:%v -> 127.0.0.1:%v", podIP, options.ProvideOptions.Expose, options.ProvideOptions.Expose)
	return nil
}
