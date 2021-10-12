package command

import (
	"context"
	"errors"
	"fmt"
	coreV1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
)

// newExchangePodCommand return new exchange command
func newExchangePodCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "exchange-pod",
		Usage: "exchange kubernetes pod to local",
		Flags: []urfave.Flag{
			urfave.StringFlag{
				Name:        "expose",
				Usage:       "ports to expose separate by comma, in [port] or [local:remote] format, e.g. 7001,8080:80",
				Destination: &options.ExchangePodOptions.Expose,
			},
			urfave.StringFlag{
				Name:        "label",
				Usage:       "the label of the pod, e.g. app=test,version=1",
				Destination: &options.ExchangePodOptions.Label,
			},
		},
		Action: func(c *urfave.Context) error {
			log.Warn().Msgf("Experimental feature. It just works on kubernetes above v1.23. It can NOT work with istio now.")
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			podToExchange := c.Args().First()
			expose := options.ExchangePodOptions.Expose

			if len(podToExchange) == 0 && len(options.ExchangePodOptions.Label) == 0 {
				return errors.New("name of pod or label of the pod to exchange is required")
			}
			if len(expose) == 0 {
				return errors.New("--expose is required")
			}
			return action.ExchangePod(podToExchange, cli, options)
		},
	}
}

// ExchangePod exchange kubernetes workload
func (action *Action) ExchangePod(podName string, cli kt.CliInterface, options *options.DaemonOptions) error {
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

	ctx := context.Background()

	var pod *coreV1.Pod
	if podName != "" {
		pod, err = kubernetes.Pod(ctx, podName, options.Namespace)
		if err != nil {
			return err
		}
	} else {
		pods, err := kubernetes.Pods(ctx, options.ExchangePodOptions.Label, options.Namespace)
		if err != nil {
			return err
		}

		for i := range pods.Items {
			pod = &pods.Items[i]
			if pod.Status.Phase != coreV1.PodRunning {
				log.Info().Msgf("Pod %s phase is %s", pod.Name, &pod.Status.Phase)
			} else {
				break
			}
		}
		if pod.Status.Phase != coreV1.PodRunning {
			return fmt.Errorf("got 0 running pod for label: %s", options.ExchangePodOptions.Label)
		}
		log.Info().Msgf("Exchange with pod: %s", pod.Name)
		podName = pod.Name
	}

	containerName := "kt-" + strings.ToLower(util.RandomString(5))

	envs := make(map[string]string)
	sshcm, err := kubernetes.AddEphemeralContainer(ctx, containerName, pod.Name, options, envs)
	if err != nil {
		return err
	}

breakLoop:
	for i := 0; i < 100; i++ {
		log.Info().Msgf("Waiting for ephemeral container: %s be ready", containerName)
		pod, err := kubernetes.Pod(ctx, pod.Name, options.Namespace)
		if err != nil {
			return err
		}
		cStats := pod.Status.EphemeralContainerStatuses
		for i := range cStats {
			if cStats[i].Name == containerName {
				if cStats[i].State.Running != nil {
					break breakLoop
				} else if cStats[i].State.Terminated != nil {
					log.Error().Msgf("Ephemeral container: %s is terminated, code: %s", containerName, cStats[i].State.Terminated.ExitCode)
				}
			}
		}
		time.Sleep(2 * time.Second)
	}

	// record data
	options.RuntimeOptions.PodName = pod.Name
	options.RuntimeOptions.SSHCM = sshcm

	shadow := connect.Create(options)
	if err = shadow.Inbound(options.ExchangePodOptions.Expose, podName, pod.Status.PodIP, nil); err != nil {
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
