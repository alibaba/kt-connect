package cluster

import (
	"time"
	"github.com/alibaba/kt-connect/pkg/kt/util"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// RemoveShadow remove shadow from cluster
func RemoveShadow(kubeConfig string, namespace string, name string) {
	client, err := GetKubernetesClient(kubeConfig)
	if err != nil {
		return
	}
	deploymentsClient := client.AppsV1().Deployments(namespace)
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// CreateShadow create shadow
func CreateShadow(
	clientset *kubernetes.Clientset,
	name string,
	labels map[string]string, 
	namespace string, 
	image string,
) (podIP string, podName string, err error) {

	localIPAddress := util.GetOutboundIP()
	log.Info().Msgf("Client address %s", localIPAddress)
	labels["remoteAddress"] = localIPAddress

	client := clientset.AppsV1().Deployments(namespace)
	deployment := generatorDeployment(namespace, name, labels, image)
	result, err := client.Create(deployment)
	log.Info().Msgf("Deploying shadow deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)

	if err != nil {
		return
	}

	pod, err := waitPodReady(namespace, name, clientset)

	if err != nil {
		return
	}
	log.Printf("Success deploy proxy deployment %s in namespace %s\n", result.GetObjectMeta().GetName(), namespace)
	podIP = pod.Status.PodIP
	podName = pod.GetObjectMeta().GetName()
	return
}

func waitPodReady(namespace string, name string, clientset *kubernetes.Clientset) (pod apiv1.Pod, err error) {
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
			log.Printf("Shadow Pod status is %s", pod.Status.Phase)
			if pod.Status.Phase == "Running" {
				break
			}
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Info().Msg("Shadow is ready.")
	return
}

func generatorDeployment(namespace string, name string, labels map[string]string, image string) *appsv1.Deployment {
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
							Name:  "standalone",
							Image: image,
						},
					},
				},
			},
		},
	}
}