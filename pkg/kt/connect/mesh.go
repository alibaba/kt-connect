package connect

import (
	"log"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// InitMesh prepare swap deployment
func (c *Connect) InitMesh() (err error) {
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.createMesh(clientset)
	return
}

func (c *Connect) createMesh(clientset *kubernetes.Clientset) (err error) {
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	meshVersion := strings.ToLower(util.RandomString(5))
	c.Name = origin.GetObjectMeta().GetName() + "-kt-" + meshVersion
	c.labels = origin.Spec.Selector.MatchLabels
	c.labels["kt"] = c.Name
	c.labels["version"] = meshVersion
	c.labels["kt-component"] = "mesh"
	c.labels["control-by"] = "kt"

	podIP, err := createAndWait(clientset, c.Namespace, c.Name, c.labels, c.Image)
	if err != nil {
		return err
	}

	log.Printf("-----------------------------------------------------------\n")
	log.Printf("|    Mesh Version '%s' You can update Istio rule       |\n", meshVersion)
	log.Printf("-----------------------------------------------------------\n")

	c.podIP = podIP
	return
}
