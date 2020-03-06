package command

import (
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/urfave/cli"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newExchangeCommand return new exchange command
func newExchangeCommand(options *options.DaemonOptions, action ActionInterface) cli.Command {
	return cli.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "expose",
				Usage:       "expose port",
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
	ch := SetUpCloseHandler(options)

	checkConnectRunning(options.RuntimeOptions.PidFile)
	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return err
	}

	origin, err := clientset.AppsV1().Deployments(options.Namespace).Get(exchange, metav1.GetOptions{})
	if err != nil {
		return err
	}

	replicas := origin.Spec.Replicas

	// Prepare context inorder to remove after command exit
	options.RuntimeOptions.Origin = exchange
	options.RuntimeOptions.Replicas = *replicas

	_, err = connect.Exchange(options, origin, clientset, util.String2Map(options.Labels))
	if err != nil {
		return err
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}
