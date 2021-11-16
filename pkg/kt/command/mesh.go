package command

import (
	"context"
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
	log.Info().Msg("---------------------------------------------------------")
	log.Info().Msgf("    Mesh Version '%s' You can update Istio rule     ", meshVersion)
	log.Info().Msg("---------------------------------------------------------")

	shadowPodName := deploymentName + "-kt-mesh-" + meshVersion
	labels := getMeshLabels(shadowPodName, meshVersion, app)
	if err = createShadowAndInbound(ctx, k, shadowPodName, labels, options); err != nil {
		return err
	}
	return nil
}

func autoMesh(ctx context.Context, k cluster.KubernetesInterface, deploymentName string, options *options.DaemonOptions) error {
	app, err := k.GetDeployment(ctx, deploymentName, options.Namespace)
	if err != nil {
		return err
	}

	svc, err := getServiceByDeployment(ctx, k, app, options)
	if err != nil {
		return err
	}

	meshVersion := getVersion(options)
	targetPorts := make([]string, 0)
	ports := make(map[int]int)
	for _, p := range svc.Spec.Ports {
		targetPorts = append(targetPorts, strconv.Itoa(p.TargetPort.IntValue()))
		ports[int(p.Port)] = p.TargetPort.IntValue()
	}

	routerPodName := deploymentName + "-kt-router"
	if err = createRouter(ctx, k, routerPodName, svc.Name, targetPorts, meshVersion, options); err != nil {
		return err
	}

	if err = createOriginService(ctx, k, svc.Name, ports, app.Spec.Selector.MatchLabels, options); err != nil {
		return err
	}

	shadowSvcName := svc.Name + "-kt-" + meshVersion
	if err = createShadowService(ctx, k, shadowSvcName, ports, app.Spec.Selector.MatchLabels, options); err != nil {
		return err
	}

	shadowPodName := deploymentName + "-kt-mesh-" + meshVersion
	labels := map[string]string{common.ControlBy: common.KubernetesTool}
	if err = createShadowAndInbound(ctx, k, shadowPodName, labels, options); err != nil {
		return err
	}
	return nil
}

func createShadowService(ctx context.Context, k cluster.KubernetesInterface, shadowSvcName string, ports map[int]int,
	selectors map[string]string, options *options.DaemonOptions) error {
	if _, err := k.CreateService(ctx, &cluster.SvcMetaAndSpec{
		Meta: &cluster.ResourceMeta{
			Name:        shadowSvcName,
			Namespace:   options.Namespace,
			Labels:      map[string]string{common.ControlBy: common.KubernetesTool},
			Annotations: map[string]string{},
		},
		External:  false,
		Ports:     ports,
		Selectors: selectors,
	}); err != nil {
		return err
	}
	log.Info().Msgf("Service %s created", shadowSvcName)
	return nil
}

func getServiceByDeployment(ctx context.Context, k cluster.KubernetesInterface, app *v1.Deployment,
	options *options.DaemonOptions) (coreV1.Service, error) {
	svcList, err := k.GetServices(ctx, app.Spec.Selector.MatchLabels, options.Namespace)
	if err != nil {
		return coreV1.Service{}, err
	} else if len(svcList) == 0 {
		return coreV1.Service{}, fmt.Errorf("failed to find service for deployment \"%s\", with labels \"%v\"",
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
	return svc, nil
}

func createRouter(ctx context.Context, k cluster.KubernetesInterface, routerPodName string, svcName string,
	targetPorts []string, meshVersion string, options *options.DaemonOptions) error {
	routerPod, err := k.GetPod(ctx, routerPodName, options.Namespace)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		labels := map[string]string{common.ControlBy: common.KubernetesTool}
		annotations := map[string]string{common.KTRefCount: "1"}
		if err = cluster.CreateRouterPod(ctx, k, routerPodName, options, labels, annotations); err != nil {
			log.Error().Err(err).Msgf("Failed to create router pod")
			return err
		}
		log.Info().Msgf("Router pod is ready")

		if _, _, err = k.ExecInPod(common.DefaultContainer, routerPodName, options.Namespace, *options.RuntimeOptions,
			"/usr/sbin/router", "setup", svcName, strings.Join(targetPorts, ","), meshVersion); err != nil {
			return err
		}
	} else {
		if _, err = strconv.Atoi(routerPod.Annotations[common.KTRefCount]); err != nil {
			log.Error().Msgf("Router pod exists, but do not have ref count")
			return err
		} else if err = k.IncreaseRef(ctx, routerPodName, options.Namespace); err != nil {
			log.Error().Msgf("Failed to increase router pod ref count")
			return err
		}
		log.Info().Msgf("Router pod already exists")

		if _, _, err = k.ExecInPod(common.DefaultContainer, routerPodName, options.Namespace, *options.RuntimeOptions,
			"/usr/sbin/router", "add", meshVersion); err != nil {
			return err
		}
	}
	log.Info().Msgf("Router pod configuration done")
	return nil
}

func createOriginService(ctx context.Context, k cluster.KubernetesInterface, svcName string,
	ports map[int]int, selectors map[string]string, options *options.DaemonOptions) error {
	originSvcName := svcName + "-kt-origin"
	_, err := k.GetService(ctx, originSvcName, options.Namespace)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		if _, err = k.CreateService(ctx, &cluster.SvcMetaAndSpec{
			Meta: &cluster.ResourceMeta{
				Name:        originSvcName,
				Namespace:   options.Namespace,
				Labels:      map[string]string{common.ControlBy: common.KubernetesTool},
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
		common.ControlBy:   common.KubernetesTool,
		common.KTComponent: common.ComponentMesh,
		common.KTName:      workload,
	}
	if meshVersion != "" {
		labels[common.KTVersion] = meshVersion
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
