package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/command/general"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfave "github.com/urfave/cli"
	"k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"os"
	"strconv"
	"strings"
)

// NewMeshCommand return new mesh command
func NewMeshCommand(cli kt.CliInterface, options *options.DaemonOptions, action ActionInterface) urfave.Command {
	return urfave.Command{
		Name:  "mesh",
		Usage: "mesh kubernetes deployment to local",
		Flags: general.MeshActionFlag(options),
		Action: func(c *urfave.Context) error {
			if options.Debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			if err := general.CombineKubeOpts(options); err != nil {
				return err
			}
			resourceToMesh := c.Args().First()
			expose := options.MeshOptions.Expose

			if len(resourceToMesh) == 0 {
				return errors.New("name of deployment to mesh is required")
			}

			if len(expose) == 0 {
				return errors.New("--expose is required")
			}

			return action.Mesh(resourceToMesh, cli, options)
		},
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(resourceName string, cli kt.CliInterface, options *options.DaemonOptions) error {
	ch, err := general.SetupProcess(cli, options, common.ComponentMesh)
	if err != nil {
		return err
	}

	if port := util.FindBrokenPort(options.MeshOptions.Expose); port != "" {
		return fmt.Errorf("no application is running on port %s", port)
	}

	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}

	ctx := context.Background()
	if options.MeshOptions.Method == common.MeshMethodManual {
		err = manualMesh(ctx, kubernetes, resourceName, options)
	} else if options.MeshOptions.Method == common.MeshMethodAuto {
		err = autoMesh(ctx, kubernetes, resourceName, options)
	} else {
		err = fmt.Errorf("invalid mesh method \"%s\"", options.MeshOptions.Method)
	}
	if err != nil {
		return err
	}

	// watch background process, clean the workspace and exit if background process occur exception
	go func() {
		<-process.Interrupt()
		log.Error().Msgf("Command interrupted")
		general.CleanupWorkspace(cli, options)
		os.Exit(0)
	}()

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

func manualMesh(ctx context.Context, k cluster.KubernetesInterface, deploymentName string, options *options.DaemonOptions) error {
	app, err := k.GetDeployment(ctx, deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	meshVersion := getVersion(options)
	shadowPodName := deploymentName + common.KtMeshInfix + meshVersion
	labels := getMeshLabels(shadowPodName, meshVersion, app)
	if err = createShadowAndInbound(ctx, k, shadowPodName, labels, options); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf(" Now you can update Istio rule by label '%s=%s' ", common.KtVersion, meshVersion)
	log.Info().Msg("---------------------------------------------------------")
	return nil
}

func autoMesh(ctx context.Context, k cluster.KubernetesInterface, deploymentName string, opts *options.DaemonOptions) error {
	app, err := k.GetDeployment(ctx, deploymentName, opts.Namespace)
	if err != nil {
		return err
	}

	svc, err := getServiceByDeployment(ctx, k, app, opts)
	if err != nil {
		return err
	}

	meshVersion := getVersion(opts)
	opts.RuntimeOptions.Mesh = meshVersion

	targetPorts := make([]string, 0)
	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		targetPorts = append(targetPorts, strconv.Itoa(p.TargetPort.IntValue()))
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	originSvcName := svc.Name + common.OriginServiceSuffix
	if err = createOriginService(ctx, k, originSvcName, ports, app.Spec.Selector.MatchLabels, opts); err != nil {
		return err
	}

	shadowPodName := deploymentName + common.KtMeshInfix + meshVersion
	shadowSvcName := svc.Name + common.KtMeshInfix + meshVersion
	shadowLabels := map[string]string{common.KtRole: "shadow", common.KtName: shadowPodName}
	if err = createShadowService(ctx, k, shadowSvcName, ports, shadowLabels, opts); err != nil {
		return err
	}

	routerPodName := deploymentName + common.RouterPodSuffix
	routerLabels := map[string]string{common.KtRole: "router", common.KtName: routerPodName}
	if err = createRouter(ctx, k, routerPodName, svc.Name, targetPorts, routerLabels, meshVersion, opts); err != nil {
		return err
	}

	if _, ok := svc.Annotations[common.KtSelector]; !ok {
		if marshaledSelector, err2 := json.Marshal(svc.Spec.Selector); err2 != nil {
			log.Error().Err(err).Msgf("Unable to record original pod selector of service %s", svc.Name)
			return err2
		} else {
			svc.Annotations[common.KtSelector] = string(marshaledSelector)
		}
	}
	svc.Spec.Selector = routerLabels
	if _, err = k.UpdateService(ctx, svc); err != nil {
		return err
	}

	if err = createShadowAndInbound(ctx, k, shadowPodName, shadowLabels, opts); err != nil {
		return err
	}
	log.Info().Msg("---------------------------------------------------------------")
	log.Info().Msgf(" Now you can access your service by header 'KT-VERSION: %s' ", meshVersion)
	log.Info().Msg("---------------------------------------------------------------")
	return nil
}

func createShadowService(ctx context.Context, k cluster.KubernetesInterface, shadowSvcName string, ports map[int]int,
	selectors map[string]string, options *options.DaemonOptions) error {
	if _, err := k.CreateService(ctx, &cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        shadowSvcName,
			Namespace:   options.Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		External:  false,
		Ports:     ports,
		Selectors: selectors,
	}); err != nil {
		return err
	}

	options.RuntimeOptions.Service = shadowSvcName
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func getServiceByDeployment(ctx context.Context, k cluster.KubernetesInterface, app *v1.Deployment,
	options *options.DaemonOptions) (*coreV1.Service, error) {
	svcList, err := k.GetServices(ctx, app.Spec.Selector.MatchLabels, options.Namespace)
	if err != nil {
		return nil, err
	} else if len(svcList) == 0 {
		return nil, fmt.Errorf("failed to find service for deployment \"%s\", with labels \"%v\"",
			app.Name, app.Spec.Selector.MatchLabels)
	} else if len(svcList) > 1 {
		svcNames := svcList[0].Name
		for i, svc := range svcList {
			if i > 0 {
				svcNames = svcNames + ", " + svc.Name
			}
		}
		log.Warn().Msgf("Found %d services match deployment \"%s\": %s. First one will be used.",
			len(svcList), app.Name, svcNames)
	}
	svc := svcList[0]
	if strings.HasSuffix(svc.Name, common.OriginServiceSuffix) {
		return k.GetService(ctx, strings.TrimSuffix(svc.Name, common.OriginServiceSuffix), options.Namespace)
	}
	return &svc, nil
}

func createRouter(ctx context.Context, k cluster.KubernetesInterface, routerPodName string, svcName string,
	targetPorts []string, routerLabels map[string]string, meshVersion string, opts *options.DaemonOptions) error {
	routerPod, err := k.GetPod(ctx, routerPodName, opts.Namespace)
	routerLabels[common.ControlBy] = common.KubernetesTool
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		annotations := map[string]string{common.KtRefCount: "1", common.KtConfig: fmt.Sprintf("service=%s", svcName)}
		if err = cluster.CreateRouterPod(ctx, k, routerPodName, opts, routerLabels, annotations); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		stdout, stderr, err2 := k.ExecInPod(common.DefaultContainer, routerPodName, opts.Namespace, *opts.RuntimeOptions,
			"/usr/sbin/router", "setup", svcName, strings.Join(targetPorts, ","), meshVersion)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	} else {
		if _, err = strconv.Atoi(routerPod.Annotations[common.KtRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = k.IncreaseRef(ctx, routerPodName, opts.Namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		stdout, stderr, err2 := k.ExecInPod(common.DefaultContainer, routerPodName, opts.Namespace, *opts.RuntimeOptions,
			"/usr/sbin/router", "add", meshVersion)
		log.Debug().Msgf("Stdout: %s", stdout)
		log.Debug().Msgf("Stderr: %s", stderr)
		if err2 != nil {
			return err2
		}
	}
	log.Info().Msgf("Router pod configuration done")
	opts.RuntimeOptions.Router = routerPodName
	return nil
}

func createOriginService(ctx context.Context, k cluster.KubernetesInterface, originSvcName string,
	ports map[int]int, selectors map[string]string, options *options.DaemonOptions) error {

	_, err := k.GetService(ctx, originSvcName, options.Namespace)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		if _, err = k.CreateService(ctx, &cluster.SvcMetaAndSpec{
			Meta: &cluster.ResourceMeta{
				Name:        originSvcName,
				Namespace:   options.Namespace,
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			External:  false,
			Ports:     ports,
			Selectors: selectors,
		}); err != nil {
			return err
		}
		log.Info().Msgf("Service %s created", originSvcName)
	} else {
		log.Info().Msgf("Origin service already exists")
	}
	return nil
}

func createShadowAndInbound(ctx context.Context, k cluster.KubernetesInterface, shadowPodName string,
	labels map[string]string, options *options.DaemonOptions) error {

	labels[common.ControlBy] = common.KubernetesTool
	envs := make(map[string]string)
	annotations := make(map[string]string)
	_, podName, sshConfigMapName, _, err := cluster.GetOrCreateShadow(ctx, k, shadowPodName, options, labels, annotations, envs)
	if err != nil {
		return err
	}

	// record context data
	options.RuntimeOptions.Shadow = shadowPodName
	options.RuntimeOptions.SSHCM = sshConfigMapName

	shadow := connect.Create(options)
	if err = shadow.Inbound(options.MeshOptions.Expose, podName); err != nil {
		return err
	}
	return nil
}

func getMeshLabels(workload string, meshVersion string, app *v1.Deployment) map[string]string {
	labels := map[string]string{
		common.KtComponent: common.ComponentMesh,
		common.KtName:      workload,
		common.KtVersion:   meshVersion,
	}
	if app != nil {
		for k, v := range app.Spec.Selector.MatchLabels {
			labels[k] = v
		}
	}
	return labels
}

func getVersion(options *options.DaemonOptions) string {
	if len(options.MeshOptions.Version) != 0 {
		return options.MeshOptions.Version
	}
	return strings.ToLower(util.RandomString(5))
}
