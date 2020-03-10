package command

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/apps/v1"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"
)

// newExchangeCommand return new exchange command
func newExchangeCommand(options *options.DaemonOptions, action ActionInterface) cli.Command {
	return cli.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "expose",
				Usage:       "expose port [port] or [remote:local]",
				Destination: &options.ExchangeOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			exchange := c.Args().First()
			expose := options.ExchangeOptions.Expose

			if len(exchange) == 0 {
				return errors.New("exchange is required")
			}
			if len(expose) == 0 {
				return errors.New("-expose is required")
			}

			return action.Exchange(exchange, options)
		},
	}
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(exchange string, options *options.DaemonOptions) error {
	generator, err := util.GetSSHGenerator()
	if err != nil {
		return err
	}

	ch := SetUpCloseHandler(options)

	checkConnectRunning(options.RuntimeOptions.PidFile)

	kubernetes, err := cluster.Create(options.KubeConfig)
	if err != nil {
		return err
	}

	app, err := kubernetes.Deployment(exchange, options.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	options.RuntimeOptions.Origin = app.GetObjectMeta().GetName()
	options.RuntimeOptions.Replicas = *app.Spec.Replicas

	sshCM := fmt.Sprintf("%s-kt-ssh-key-%s", app.GetName(), strings.ToLower(util.RandomString(5)))
	workload := app.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	if _, err = kubernetes.CreateSSHCM(sshCM,
		options.Namespace,
		getExchangeLables(options.Labels, sshCM, app),
		map[string]string{
			"key": string(generator.PublicKey),
		},
	); err != nil {
		return err
	}

	podIP, podName, err := kubernetes.CreateShadow(
		workload, options.Namespace, options.Image, sshCM, getExchangeLables(options.Labels, workload, app))
	log.Info().Msgf("create exchange shadow %s in namespace %s", workload, options.Namespace)

	if err != nil {
		return err
	}

	// record data
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshCM

	down := int32(0)
	kubernetes.Scale(app, &down)

	shadow := connect.Create(options)
	shadow.Inbound(options.ExchangeOptions.Expose, podName, podIP)

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func getExchangeLables(customLabels string, workload string, origin *v1.Deployment) map[string]string {
	labels := map[string]string{
		"kt":           workload,
		"kt-component": "exchange",
		"control-by":   "kt",
	}
	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}
	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(customLabels) {
		labels[k] = v
	}
	return labels
}
