package connect

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Exchange exchange request to local
func (c *Connect) Exchange(options *options.DaemonOptions, origin *v1.Deployment, clientset *kubernetes.Clientset,
	labels map[string]string) (workload string, err error) {
	workload, podIP, podName, err := c.createExchangeShadow(origin, options.Namespace, clientset, labels)
	options.RuntimeOptions.Shadow = workload
	down := int32(0)
	scaleTo(origin, options.Namespace, clientset, &down)
	remotePortForward(c.Expose, c.Kubeconfig, options.Namespace, podName, podIP, c.Debug)
	return
}

// HandleExchangeExit handle error when exchange exit
func (c *Connect) HandleExchangeExit(shadow string, replicas *int32, origin *v1.Deployment, clientset *kubernetes.Clientset) {
	log.Printf("Cleanup proxy shadow %s", shadow)
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient.Delete(shadow, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	app, err := clientset.AppsV1().Deployments(c.Namespace).Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return
	}

	scaleTo(app, c.Namespace, clientset, replicas)
}

//ScaleTo Scale
func scaleTo(deployment *v1.Deployment, namespace string, clientset *kubernetes.Clientset, replicas *int32) (err error) {
	log.Printf("Try Scale deployment %s to %d\n", deployment.GetObjectMeta().GetName(), *replicas)
	client := clientset.AppsV1().Deployments(namespace)
	deployment.Spec.Replicas = replicas

	d, err := client.Update(deployment)
	if err != nil {
		log.Printf("%s Fails scale deployment %s to %d\n", err.Error(), deployment.GetObjectMeta().GetName(), *replicas)
		return err
	}
	log.Printf(" * %s (%d replicas) success", d.Name, *d.Spec.Replicas)
	return nil
}

func (c *Connect) createExchangeShadow(origin *v1.Deployment, namespace string, clientset *kubernetes.Clientset,
	extraLabels map[string]string) (workload string, podIP string, podName string, err error) {
	workload = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	labels := map[string]string{
		"kt":           workload,
		"kt-component": "exchange",
		"control-by":   "kt",
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}

	podIP, podName, err = cluster.CreateShadow(clientset, workload, labels, c.Image, namespace)
	if err != nil {
		return "", "", "", err
	}
	return
}
