package connect

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Exchange
func (c *Connect) Exchange(namespace string, origin *v1.Deployment, clientset *kubernetes.Clientset) (shadow string, err error) {
	shadow, podIP, err := c.createShadow(origin, namespace, clientset)
	down := int32(0)
	scaleTo(origin, namespace, clientset, &down)
	remotePortForward(c.Expose, c.Kubeconfig, c.Namespace, shadow, podIP, c.Debug)
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

func remotePortForward(expose string, kubeconfig string, namespace string, target string, remoteIP string, debug bool) (err error) {
	localSSHPort, err := strconv.Atoi(util.GetRandomSSHPort(remoteIP))
	if err != nil {
		return
	}
	portforward := util.PortForward(kubeconfig, namespace, target, localSSHPort)
	err = util.BackgroundRun(portforward, "exchange port forward to local", debug)
	if err != nil {
		return
	}

	time.Sleep(time.Duration(2) * time.Second)
	log.Printf("SSH Remote port-forward POD %s 22 to 127.0.0.1:%d starting\n", remoteIP, localSSHPort)
	cmd := util.SSHRemotePortForward(expose, "127.0.0.1", expose, localSSHPort)
	return util.BackgroundRun(cmd, "ssh remote port-forward", debug)
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

func (c *Connect) createShadow(origin *v1.Deployment, namespace string, clientset *kubernetes.Clientset) (shadowName string, podIP string, err error) {
	shadowName = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))
	c.swapReplicas = origin.Spec.Replicas
	c.Name = shadowName

	labels := map[string]string{
		"kt":           shadowName,
		"kt-component": "exchange",
		"control-by":   "kt",
	}

	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}

	podIP, err = createAndWait(clientset, namespace, shadowName, labels, c.Image)
	if err != nil {
		return "", "", err
	}
	return
}
