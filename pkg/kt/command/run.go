package command

import (
	"errors"
	"strconv"
	"strings"

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
				Usage:       " If true, a public, external service is created",
				Destination: &options.RunOptions.Expose,
			},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			port := options.RunOptions.Port
			if port == 0 {
				return errors.New("--port is required")
			}
			return action.Run(c.Args().First(), cli, options)
		},
	}
}

// Run create a new service in cluster
func (action *Action) Run(service string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch := SetUpCloseHandler(options)
	run(service, cli, options)
	<-ch
	return nil
}

// Run create a new service in cluster
func run(service string, cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	labels := map[string]string{
		"control-by":   "kt",
		"kt-component": "run",
		"kt":           service,
		"version":      strings.ToLower(util.RandomString(5)),
	}

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	podIP, podName, sshcm, credential, err := kubernetes.CreateShadow(service, options.Namespace, options.Image, labels, options.Debug)
	if err != nil {
		return err
	}
	log.Info().Msgf("create shadow pod %s ip %s", podName, podIP)

	if options.RunOptions.Expose {
		log.Info().Msgf("expose deployment %s to %s:%v", service, service, options.RunOptions.Port)
		_, err = kubernetes.CreateService(service, options.Namespace, options.RunOptions.Port, labels)
		if err != nil {
			return err
		}
		options.RuntimeOptions.Service = service
	}

	options.RuntimeOptions.Shadow = service
	options.RuntimeOptions.SSHCM = sshcm

	err = cli.Shadow().Inbound(strconv.Itoa(options.RunOptions.Port), podName, podIP, credential)
	if err != nil {
		return err
	}

	log.Info().Msgf("forward remote %s:%v -> 127.0.0.1:%v", podIP, options.RunOptions.Port, options.RunOptions.Port)
	return nil
}
