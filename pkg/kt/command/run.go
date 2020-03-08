package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"
)

// newRunCommand return new run command
func newRunCommand(options *options.DaemonOptions, action ActionInterface) cli.Command {
	return cli.Command{
		Name:  "run",
		Usage: "create a shadow deployment to redirect request to user local",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:        "port",
				Usage:       "The port that exposes",
				Destination: &options.RunOptions.Port,
			},
			cli.BoolFlag{
				Name:        "expose",
				Usage:       " If true, a public, external service is created",
				Destination: &options.RunOptions.Expose,
			},
		},
		Action: func(c *cli.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			port := options.RunOptions.Port
			if port == 0 {
				return errors.New("--port is required")
			}
			return action.Run(c.Args().First(), options)
		},
	}
}

// Run create a new service in cluster
func (action *Action) Run(service string, options *options.DaemonOptions) error {
	generator, err := util.GetSSHGenerator()
	if err != nil {
		return err
	}
	ch := SetUpCloseHandler(options)
	kubernetes, err := cluster.Create(options.KubeConfig)
	if err != nil {
		return err
	}

	labels := map[string]string{
		"control-by":   "kt",
		"kt-component": "run",
	}

	// extra labels must be applied after origin labels
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	sshCM := fmt.Sprintf("kt-ssh-key-%s", strings.ToLower(util.RandomString(5)))
	labels["kt"] = sshCM
	if _, err = kubernetes.CreateSSHCM(sshCM,
		options.Namespace,
		labels,
		map[string]string{
			"key": string(generator.PublicKey),
		},
	); err != nil {
		return err
	}
	labels["kt"] = service
	podIP, podName, err := kubernetes.CreateShadow(service, options.Namespace, options.Image, sshCM, labels)
	if err != nil {
		return err
	}
	log.Info().Msgf("create shadow pod %s ip %s", podName, podIP)

	if options.RunOptions.Expose {
		log.Info().Msgf("expose deployment %s to %s:%v", service, service, options.RunOptions.Port)
		if _, err = kubernetes.CreateService(service,
			options.Namespace,
			labels,
			options.RunOptions.Port,
		); err != nil {
			return err
		}
		options.RuntimeOptions.Service = service
	}

	options.RuntimeOptions.Shadow = service
	options.RuntimeOptions.SSHCM = sshCM

	shadow := connect.Create(options)
	err = shadow.Inbound(strconv.Itoa(options.RunOptions.Port), podName, podIP)
	if err != nil {
		return err
	}

	log.Info().Msgf("forward remote %s:%v -> 127.0.0.1:%v", podIP, options.RunOptions.Port, options.RunOptions.Port)

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}
