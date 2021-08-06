package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/kt-connect/internal/process"
	"github.com/alibaba/kt-connect/pkg/common"

	"github.com/alibaba/kt-connect/pkg/kt"

	v1 "k8s.io/api/apps/v1"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	urfave "github.com/urfave/cli"
)

// newExchangeCommand return new exchange command
func newExchangeCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: []urfave.Flag{
			urfave.StringFlag{
				Name:        "expose",
				Usage:       "expose port [port] or [remote:local]",
				Destination: &options.ExchangeOptions.Expose,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			deploymentToExchange := c.Args().First()
			expose := options.ExchangeOptions.Expose

			if len(deploymentToExchange) == 0 {
				return errors.New("name of deployment to exchange is required")
			}
			if len(expose) == 0 {
				return errors.New("--expose is required")
			}
			return action.Exchange(deploymentToExchange, cli, options)
		},
	}
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	options.RuntimeOptions.Component = common.ComponentExchange
	err := util.WritePidFile(common.ComponentExchange)
	if err != nil {
		return err
	}
	log.Info().Msgf("KtConnect start at %d", os.Getpid())

	ch := SetUpCloseHandler(cli, options, common.ComponentExchange)

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	app, err := kubernetes.Deployment(deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	options.RuntimeOptions.Origin = app.GetName()
	options.RuntimeOptions.Replicas = *app.Spec.Replicas

	workload := app.GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	envs := make(map[string]string)
	podIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(workload, options,
		getExchangeLabels(options, workload, app), getExchangeAnnotation(options), envs)
	log.Info().Msgf("Create exchange shadow %s in namespace %s", workload, options.Namespace)

	if err != nil {
		return err
	}

	// record data
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	down := int32(0)
	if err = kubernetes.Scale(app, &down); err != nil {
		return err
	}

	shadow := connect.Create(options)
	if err = shadow.Inbound(options.ExchangeOptions.Expose, podName, podIP, credential); err != nil {
		return err
	}

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted: %s", <-process.Interrupt())
		CleanupWorkspace(cli, options)
		os.Exit(0)
	}()
	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func getExchangeAnnotation(options *options.DaemonOptions) map[string]string {
	return map[string]string{
		common.KTConfig: fmt.Sprintf("app=%s,replicas=%d",
			options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas),
	}
}

func getExchangeLabels(options *options.DaemonOptions, workload string, origin *v1.Deployment) map[string]string {
	labels := map[string]string{
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentExchange,
		common.KTName:      workload,
	}
	if origin != nil {
		for k, v := range origin.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}
	splits := strings.Split(workload, "-")
	labels[common.KTVersion] = splits[len(splits)-1]
	return labels
}
