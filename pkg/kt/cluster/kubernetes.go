package cluster

import (
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	clusterWatcher "github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

// Scale scale deployment to
func (k *Kubernetes) Scale(deployment *appV1.Deployment, replicas *int32) (err error) {
	log.Info().Msgf("scale deployment %s to %d\n", deployment.GetObjectMeta().GetName(), *replicas)
	client := k.Clientset.AppsV1().Deployments(deployment.GetObjectMeta().GetNamespace())
	deployment.Spec.Replicas = replicas

	d, err := client.Update(deployment)
	if err != nil {
		log.Error().Msgf("%s Fails scale deployment %s to %d\n", err.Error(), deployment.GetObjectMeta().GetName(), *replicas)
		return
	}
	log.Info().Msgf(" * %s (%d replicas) success", d.Name, *d.Spec.Replicas)
	return
}

// Deployment get deployment
func (k *Kubernetes) Deployment(name, namespace string) (deployment *appV1.Deployment, err error) {
	deployment, err = k.Clientset.AppsV1().Deployments(namespace).Get(name, metaV1.GetOptions{})
	return
}

// CreateShadow create shadow
func (k *Kubernetes) CreateShadow(name, namespace, image string, labels map[string]string) (podIP, podName string, err error) {
	return CreateShadow(k.Clientset, name, labels, namespace, image)
}

// ClusterCrids get cluster cirds
func (k *Kubernetes) ClusterCrids(podCIDR string) (cidrs []string, err error) {
	serviceList, err := k.Clientset.CoreV1().Services("").List(metaV1.ListOptions{})
	if err != nil {
		return
	}

	cidrs, err = getPodCirds(k.Clientset, podCIDR)
	if err != nil {
		return
	}

	services := serviceList.Items
	serviceCird, err := getServiceCird(services)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCird...)
	return
}

// ServiceHosts get service dns map
func (k *Kubernetes) ServiceHosts(namespace string) (hosts map[string]string) {
	services, err := k.Clientset.CoreV1().Services(namespace).List(metaV1.ListOptions{})
	if err != nil {
		return
	}
	hosts = map[string]string{}
	for _, service := range services.Items {
		hosts[service.ObjectMeta.Name] = service.Spec.ClusterIP
	}
	return
}

// ScaleTo scale app
func ScaleTo(clientSet *kubernetes.Clientset, namespace, name string, replicas int32) (err error) {
	client := clientSet.AppsV1().Deployments(namespace)
	deployment, err := client.Get(name, metaV1.GetOptions{})
	if err != nil {
		return
	}

	// make sure min replicas
	if replicas == 0 {
		replicas = 1
	}

	log.Info().Msgf("- Scale %s in %s to %d", name, namespace, replicas)

	deployment.Spec.Replicas = &replicas
	_, err = client.Update(deployment)
	return
}

// RemoveShadow remove shadow from cluster
func RemoveShadow(client *kubernetes.Clientset, namespace, name string) {
	deploymentsClient := client.AppsV1().Deployments(namespace)
	deletePolicy := metaV1.DeletePropagationBackground
	err := deploymentsClient.Delete(name, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Error().Err(err).Msgf("delete deployment %s failed", name)
	}
}

// CreateService create service in cluster
func CreateService(name, namespace string,
	labels map[string]string,
	port int,
	clientset *kubernetes.Clientset,
) (err error) {
	client := clientset.CoreV1().Services(namespace)
	svc := generateService(name, namespace, labels, port)
	_, err = client.Create(svc)
	return err
}

// RemoveService create service in cluster
func RemoveService(
	name, namespace string,
	clientset *kubernetes.Clientset,
) (err error) {
	client := clientset.CoreV1().Services(namespace)
	return client.Delete(name, &metaV1.DeleteOptions{})
}

// CreateShadow create shadow
func CreateShadow(
	clientset *kubernetes.Clientset,
	name string,
	labels map[string]string,
	namespace,
	image string,
) (podIP, podName string, err error) {

	localIPAddress := util.GetOutboundIP()
	log.Info().Msgf("Client address %s", localIPAddress)
	labels["remoteAddress"] = localIPAddress

	client := clientset.AppsV1().Deployments(namespace)
	deployment := generatorDeployment(namespace, name, labels, image)
	result, err := client.Create(deployment)
	if err != nil {
		return
	}
	log.Info().Msgf("deploy shadow deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)

	// pod, err := waitPodReady(namespace, name, clientset)
	pod, err := waitPodReadyUsingInformer(namespace, name, clientset)
	if err != nil {
		return
	}
	podIP = pod.Status.PodIP
	podName = pod.GetObjectMeta().GetName()
	return
}

func waitPodReadyUsingInformer(namespace, name string, clientset *kubernetes.Clientset) (pod v1.Pod, err error) {
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	podListener, err := clusterWatcher.PodListener(clientset, stopSignal)
	if err != nil {
		return
	}
	pod = v1.Pod{}
	podLabels := labels.NewSelector()
	log.Info().Msgf("pod label: kt=%s", name)
	labelKeys := []string{
		"kt",
	}
	requirement, err := labels.NewRequirement(labelKeys[0], selection.Equals, []string{name})
	if err != nil {
		return
	}
	podLabels.Add(*requirement)

	pods, err := podListener.Pods(namespace).List(podLabels)
	if err != nil {
		return pod, err
	}
	getTargetPod := func(podList []*v1.Pod) *v1.Pod {
		// log.Info().Msgf("len(podList):%d", len(podList))
		for _, p := range podList {
			if len(p.Labels) <= 0 {
				// almost impossible
				continue
			}
			item, containKey := p.Labels[labelKeys[0]]
			if !containKey || item != name {
				continue
			}
			return p
		}
		return nil
	}
	wait := func(podName string) {
		time.Sleep(time.Second)
		if len(podName) >= 0 {
			log.Info().Msgf("pod: %s is running,but not ready", podName)
			return
		}
		log.Info().Msg("Shadow Pods not ready......")
	}
wait_loop:
	for {
		hasRunningPod := len(pods) > 0
		var podName string
		if hasRunningPod {
			// podLister do not support FieldSelector
			// https://github.com/kubernetes/client-go/issues/604
			p := getTargetPod(pods)
			if p != nil {
				if p.Status.Phase == "Running" {
					pod = *p
					log.Info().Msgf("Shadow pod: %s is ready.", pod.Name)
					// stop watcher
					stopSignal <- struct{}{}
					break wait_loop
				}
				podName = p.Name
			}
		}
		wait(podName)
		pods, err = podListener.Pods(namespace).List(podLabels)
		if err != nil {
			return pod, err
		}
	}
	return pod, nil
}

func generateService(name, namespace string, labels map[string]string, port int) *v1.Service {
	var ports []v1.ServicePort
	ports = append(ports, v1.ServicePort{
		Name:       name,
		Port:       int32(port),
		TargetPort: intstr.FromInt(port),
	})

	return &v1.Service{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}

}

func generatorDeployment(namespace, name string, labels map[string]string, image string) *appV1.Deployment {
	return &appV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appV1.DeploymentSpec{
			Selector: &metaV1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "standalone",
							Image:           image,
							ImagePullPolicy: "Always",
						},
					},
				},
			},
		},
	}
}
