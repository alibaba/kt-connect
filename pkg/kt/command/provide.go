package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// NewProvideCommand return new provide command
func NewProvideCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "provide",
		Usage: "create a shadow service to redirect request to user local",
		Flags: general.ProvideActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
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
	ch, err := general.SetupProcess(cli, options, common.ComponentProvide)
	if err != nil {
		return err
	}

	if port := util.FindBrokenPort(strconv.Itoa(options.ProvideOptions.Expose)); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	if err := provide(context.TODO(), serviceName, cli, options); err != nil {
		return err
	}
	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	<-ch
	return nil
}

// Provide create a new service in cluster
func provide(ctx context.Context, serviceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	version := strings.ToLower(util.RandomString(5))
	shadowPodName := fmt.Sprintf("%s-kt-%s", serviceName, version)
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentProvide,
		common.KTName:      shadowPodName,
		common.KTVersion:   version,
	}
	annotations := map[string]string{
		common.KTConfig: fmt.Sprintf("service=%s", serviceName),
	}

	return exposeLocalService(ctx, serviceName, shadowPodName, labels, annotations, options, kubernetes, cli)
}

// exposeLocalService create shadow and expose service if need
func exposeLocalService(ctx context.Context, serviceName, shadowPodName string, labels, annotations map[string]string,
	options *options.DaemonOptions, kubernetes cluster.KubernetesInterface, cli kt.CliInterface) (err error) {

	envs := make(map[string]string)
	_, podName, sshConfigMapName, _, err := cluster.GetOrCreateShadow(ctx, kubernetes, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}
	log.Info().Msgf("Created shadow pod %s", podName)

	log.Info().Msgf("Expose deployment %s to service %s:%v", shadowPodName, serviceName, options.ProvideOptions.Expose)
	_, err = kubernetes.CreateService(ctx, serviceName, options.Namespace, options.ProvideOptions.External, options.ProvideOptions.Expose, labels)
	if err != nil {
		return err
	}
	options.RuntimeOptions.Service = serviceName

	options.RuntimeOptions.Shadow = shadowPodName
	options.RuntimeOptions.SSHCM = sshConfigMapName

	err = cli.Shadow().Inbound(strconv.Itoa(options.ProvideOptions.Expose), podName)
	if err != nil {
		return err
	}

	log.Info().Msgf("Forward remote %s:%v -> 127.0.0.1:%v", podName, options.ProvideOptions.Expose, options.ProvideOptions.Expose)
	return nil
}
