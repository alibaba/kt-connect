package birdseye

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"strings"
)

const UnknownUser = "unknown user"

func GetKtPodsAndAllServices(namespace string) ([]coreV1.Pod, []appV1.Deployment, []coreV1.Service, []coreV1.Service, error) {
	pods, err := cluster.Ins().GetPodsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	apps, err := cluster.Ins().GetDeploymentsByLabel(map[string]string{util.ControlBy: util.KubernetesToolkit}, namespace)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	svcs, err := cluster.Ins().GetAllServiceInNamespace(opt.Get().Global.Namespace)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ktSvcs := make([]coreV1.Service, 0)
	otherSvcs := make([]coreV1.Service, 0)
	for _, svc := range svcs.Items {
		if svc.Labels[util.ControlBy] == util.KubernetesToolkit {
			ktSvcs = append(ktSvcs, svc)
		} else {
			otherSvcs = append(otherSvcs, svc)
		}
	}
	return pods.Items, apps.Items, ktSvcs, otherSvcs, nil
}

func GetConnectors(pods []coreV1.Pod, apps []appV1.Deployment) []string {
	users:= make([]string, 0)
	for _, pod := range pods {
		if user := checkConnector(pod.Labels, pod.Annotations); user != "" {
			users = append(users, user)
		}
	}
	for _, app := range apps {
		if user := checkConnector(app.Labels, app.Annotations); user != "" {
			users = append(users, user)
		}
	}
	return users
}

func GetServiceStatus(ktSvcs []coreV1.Service, pods []coreV1.Pod, svcs []coreV1.Service) [][]string {
	allServices := make([][]string, 0)
	for _, svc := range ktSvcs {
		for _, p := range pods {
			if p.Labels[util.KtRole] == util.RolePreviewShadow && util.MapContains(svc.Spec.Selector, p.Labels) {
				allServices = append(allServices, []string{svc.Name, "previewing by " + getUserName(p)})
				break
			}
		}
	}
	svcLoop:
	for _, svc := range svcs {
		for _, p := range pods {
			if util.MapContains(svc.Spec.Selector, p.Labels) {
				if role := p.Labels[util.KtRole]; role == util.RoleExchangeShadow {
					allServices = append(allServices, []string{svc.Name, "exchanged by " + getUserName(p)})
					continue svcLoop
				} else if role == util.RoleRouter {
					allServices = append(allServices, []string{svc.Name, "meshed (auto) by " +
						getMeshedUserNames(ktSvcs, pods, svc.Name + util.MeshPodInfix)})
					continue svcLoop
				} else if role == util.RoleMeshShadow {
					allServices = append(allServices, []string{svc.Name, "meshed (manual) by " +
						getMeshedUserNames([]coreV1.Service{svc}, pods, svc.Name)})
					continue svcLoop
				}
			}
		}
		if !opt.Get().Birdseye.HideNaturalService {
			allServices = append(allServices, []string{svc.Name, "normal"})
		}
	}
	return allServices
}

func getUserName(p coreV1.Pod) string {
	user := p.Annotations[util.KtUser]
	if user == "" {
		user = UnknownUser
	}
	return user
}

func getMeshedUserNames(svcs []coreV1.Service, pods []coreV1.Pod, namePrefix string) string {
	users := make([]string, 0)
	for _, s := range svcs {
		if strings.HasPrefix(s.Name, namePrefix) {
			for _, p := range pods {
				if p.Labels[util.KtRole] == util.RoleMeshShadow && util.MapContains(s.Spec.Selector, p.Labels) {
					user := p.Annotations[util.KtUser]
					if user != "" {
						users = append(users, user)
					}
					break
				}
			}
		}
	}
	if len(users) == 0 {
		return UnknownUser
	}
	return strings.Join(users, ", ")
}

func checkConnector(labels map[string]string, annotations map[string]string) string {
	if role, exists := labels[util.KtRole]; !exists || role != util.RoleConnectShadow {
		return ""
	}
	if user, exists := annotations[util.KtUser]; exists {
		lastHeartBeat := util.ParseTimestamp(annotations[util.KtLastHeartBeat])
		if lastHeartBeat > 0 {
			lastActiveInMin := (util.GetTime() - lastHeartBeat) / 60
			return fmt.Sprintf("%s (last active %d min ago)", user, lastActiveInMin)
		} else {
			return user
		}
	} else {
		return UnknownUser
	}
}
