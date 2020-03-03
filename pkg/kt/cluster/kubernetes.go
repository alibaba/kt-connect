package cluster

import (
	"errors"
	"time"

	clusterWatcher "github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient get Kubernetes client from config
func GetKubernetesClient(kubeConfig string) (clientset *kubernetes.Clientset, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}
	clientset, err = kubernetes.NewForConfig(config)
	return
}

// ScaleTo scale app
func ScaleTo(clientSet *kubernetes.Clientset, namespace, name string, replicas int32) (err error) {
	client := clientSet.AppsV1().Deployments(namespace)
	deployment, err := client.Get(name, metav1.GetOptions{})
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

// Remove remove shadow from cluster
func Remove(client *kubernetes.Clientset, namespace, name string) {
	deploymentsClient := client.AppsV1().Deployments(namespace)
	deletePolicy := metav1.DeletePropagationForeground
	err := deploymentsClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		log.Error().Err(err).Msgf("delete deployment %s failed", name)
	}
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
	log.Info().Msgf("Deploying shadow deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)

	// pod, err := waitPodReady(namespace, name, clientset)
	pod, err := waitPodReadyUsingInformer(namespace, name, clientset)
	if err != nil {
		return
	}
	log.Info().Msgf("Success deploy proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)
	podIP = pod.Status.PodIP
	podName = pod.GetObjectMeta().GetName()
	return
}

// Deprecated : this implement is not "golang style"
func waitPodReady(namespace, name string, clientset *kubernetes.Clientset) (pod apiv1.Pod, err error) {
	pod = apiv1.Pod{}
	for {
		pods, err1 := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: "kt=" + name,
		})

		if err1 != nil {
			err = err1
			return
		}

		if len(pods.Items) <= 0 {
			log.Printf("Shadow Pods not ready......")
		} else {
			pod = pods.Items[0]
			log.Info().Msgf("Shadow Pod status is %s", pod.Status.Phase)
			if pod.Status.Phase == "Running" {
				break
			}
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Info().Msg("Shadow is ready.")
	return
}

func waitPodReadyUsingInformer(namespace, name string, clientset *kubernetes.Clientset) (pod apiv1.Pod, err error) {
	pod = apiv1.Pod{}
	// sharedInformerFactory:=informers.SharedInformerOption()
	option := informers.WithNamespace(namespace)
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, option)
	podInformer := factory.Core().V1().Pods()
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
	// Kubernetes serves an utility to handle API crashes
	defer runtime.HandleCrash()
	stopSignal := make(chan struct{})
	defer close(stopSignal)
	go factory.Start(stopSignal)
	if !cache.WaitForNamedCacheSync("attach detach", stopSignal,
		podInformer.Informer().HasSynced) {
		err = errors.New("Error waiting for the informer caches to sync")
		if err != nil {
			return pod, err
		}
	}
	pods, err := podInformer.Lister().Pods(namespace).List(podLabels)
	if err != nil {
		return pod, err
	}
	log.Info().Msg("podInformer init")
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
					break wait_loop
				}
				podName = p.Name
			}
		}
		wait(podName)
		pods, err = podInformer.Lister().Pods(namespace).List(podLabels)
		if err != nil {
			return pod, err
		}
	}
	<-stopSignal
	return pod, nil
}

func generatorDeployment(namespace, name string, labels map[string]string, image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
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

// LocalHosts LocalHosts
func LocalHosts(clientset *kubernetes.Clientset, namespace string) (hosts map[string]string) {
	serviceListener, err := clusterWatcher.ServiceListener(clientset)
	if err != nil {
		return
	}

	services, err := serviceListener.Services(namespace).List(labels.Everything())
	if err != nil {
		return
	}

	hosts = map[string]string{}
	for _, service := range services {
		hosts[service.ObjectMeta.Name] = service.Spec.ClusterIP
	}
	return
}
