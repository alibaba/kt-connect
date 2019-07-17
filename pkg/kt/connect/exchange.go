package connect

import (
	"log"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// InitExchange prepare swap deployment
func (c *Connect) InitExchange() (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.createSwap(clientset)
	return
}

func (c *Connect) createSwap(clientset *kubernetes.Clientset) (err error) {
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	c.swapReplicas = origin.Spec.Replicas

	replicas := int32(0)
	origin.Spec.Replicas = &replicas
	d, err := deploymentsClient.Update(origin)
	if err != nil {
		log.Printf("Fails scale deployment %s to zero\n", origin.GetObjectMeta().GetName())
		return err
	}
	log.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	log.Printf("Scale deployment %s to zero\n", origin.GetObjectMeta().GetName())

	c.Name = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))

	c.labels = origin.Spec.Selector.MatchLabels
	c.labels["kt"] = c.Name
	c.labels["kt-component"] = "exchange"
	c.labels["control-by"] = "kt"

	podIP, err := createAndWait(clientset, c.Namespace, c.Name, c.labels, c.Image)
	if err != nil {
		return err
	}
	c.podIP = podIP
	return
}
