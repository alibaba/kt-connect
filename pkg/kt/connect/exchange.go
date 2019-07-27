package connect

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Exchange
func (c *Connect) Exchange(namespace string, origin *v1.Deployment, clientset *kubernetes.Clientset) (workload string, err error) {
	workload, podIP, podName, err := c.createExchangeShadow(origin, namespace, clientset)
	down := int32(0)
	scaleTo(origin, namespace, clientset, &down)
	remotePortForward(c.Expose, c.Kubeconfig, c.Namespace, podName, podIP, c.Debug)
	return
}

// HandleExchangeExit
func (c *Connect) HandleExchangeExit(shadow string, replicas *int32, origin *v1.Deployment, clientset *kubernetes.Clientset) {
	os.Remove(c.PidFile)
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

func (c *Connect) createExchangeShadow(origin *v1.Deployment, namespace string, clientset *kubernetes.Clientset) (workload string, podIP string, podName string, err error) {
	workload = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	labels := map[string]string{
		"kt":           workload,
		"kt-component": "exchange",
		"control-by":   "kt",
	}

	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}

	podIP, podName, err = createAndWait(clientset, namespace, workload, labels, c.Image)
	if err != nil {
		return "", "", "", err
	}
	return
}
