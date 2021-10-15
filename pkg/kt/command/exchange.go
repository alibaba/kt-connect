package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"time"
)

// newExchangeCommand return new exchange command
func newExchangeCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "exchange",
		Usage: "exchange kubernetes deployment to local",
		Flags: []urfave.Flag{
			urfave.StringFlag{
				Name:        "expose",
				Usage:       "ports to expose separate by comma, in [port] or [local:remote] format, e.g. 7001,8080:80",
				Destination: &options.ExchangeOptions.Expose,
			},
			urfave.StringFlag{
				Name:        "method",
				Value:       "scale",
				Usage:       "Exchange method 'scale' or 'ephemeral'",
				Destination: &options.ExchangeOptions.Method,
			},
			// TODO: should be replace with service name
			//urfave.StringFlag{
			//	Name:        "label",
			//	Usage:       "(ephemeral mode only) Label of the pod, e.g. app=test,version=1",
			//	Destination: &options.ExchangePodOptions.Label,
			//},
		},
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := combineKubeOpts(options); err != nil {
				return err
			}
			resourceToExchange := c.Args().First()
			expose := options.ExchangeOptions.Expose

			if len(resourceToExchange) == 0 {
				return errors.New("name of resource to exchange is required")
			}
			if len(expose) == 0 {
				return errors.New("--expose is required")
			}

			return action.Exchange(resourceToExchange, cli, options)
		},
	}
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(resourceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	options.RuntimeOptions.Component = common.ComponentExchange
	err := util.WritePidFile(common.ComponentExchange)
	if err != nil {
		return err
	}
	log.Info().Msgf("KtConnect %s start at %d", options.Version, os.Getpid())

	method := options.ExchangeOptions.Method
	if method == common.ExchangeMethodScale {
		return exchangeByScale(resourceName, cli, options)
	} else if method == common.ExchangeMethodEphemeral {
		return exchangeByEphemeralContainer(resourceName, cli, options)
	} else {
		return fmt.Errorf("invalid exchange method \"%s\"", method)
	}
}

func exchangeByScale(deploymentName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch := SetUpCloseHandler(cli, options, common.ComponentExchange)
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}
	ctx := context.Background()
	app, err := kubernetes.Deployment(ctx, deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	// record context inorder to remove after command exit
	options.RuntimeOptions.Origin = app.GetName()
	options.RuntimeOptions.Replicas = *app.Spec.Replicas

	workload := app.GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	envs := make(map[string]string)
	podIP, podName, sshcm, credential, err := kubernetes.GetOrCreateShadow(ctx, workload, options,
		getExchangeLabels(options, workload, app), getExchangeAnnotation(options), envs)
	log.Info().Msgf("Create exchange shadow %s in namespace %s", workload, options.Namespace)

	if err != nil {
		return err
	}

	// record data
	options.RuntimeOptions.Shadow = workload
	options.RuntimeOptions.SSHCM = sshcm

	down := int32(0)
	if err = kubernetes.Scale(ctx, app, &down); err != nil {
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
	log.Info().Msgf("Terminal signal is %s", s)

	return nil
}

func exchangeByEphemeralContainer(podName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	log.Warn().Msgf("Experimental feature. It just works on kubernetes above v1.23. It can NOT work with istio now.")
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
				log.Info().Msgf("Pod %s phase is %s", pod.Name, pod.Status.Phase)
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
					log.Error().Msgf("Ephemeral container: %s is terminated, code: %d", containerName, cStats[i].State.Terminated.ExitCode)
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

func getExchangeAnnotation(options *options.DaemonOptions) map[string]string {
	return map[string]string{
		common.KTConfig: fmt.Sprintf("app=%s,replicas=%d",
			options.RuntimeOptions.Origin, options.RuntimeOptions.Replicas),
	}
}

func getExchangeLabels(options *options.DaemonOptions, workload string, origin *appV1.Deployment) map[string]string {
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
	splits := strings.Split(workload, "-")
	labels[common.KTVersion] = splits[len(splits)-1]
	return labels
}
