package connect

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Mesh prepare swap deployment
func (c *Connect) Mesh(options *options.DaemonOptions, clientset *kubernetes.Clientset, labels map[string]string) (workload string, err error) {
	workload, podIP, podName, err := c.createMeshShadown(clientset, labels)
	options.RuntimeOptions.Shadow=workload
	remotePortForward(c.Expose, c.Kubeconfig, c.Namespace, podName, podIP, c.Debug)
	return
}

func (c *Connect) createMeshShadown(clientset *kubernetes.Clientset,
	extraLabels map[string]string) (shadowName string, podIP string, podName string, err error) {
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	origin, err := deploymentsClient.Get(c.Swap, metav1.GetOptions{})
	if err != nil {
		return "", "", "", err
	}

	meshVersion := strings.ToLower(util.RandomString(5))
	shadowName = origin.GetObjectMeta().GetName() + "-kt-" + meshVersion
	labels := map[string]string{
		"kt":           shadowName,
		"kt-component": "mesh",
		"control-by":   "kt",
		"version":      meshVersion,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}
	for k, v := range origin.Spec.Selector.MatchLabels {
		labels[k] = v
	}

	podIP, podName, err = cluster.CreateShadow(clientset, c.Namespace, shadowName, labels, c.Image)
	if err != nil {
		return "", "", "", err
	}

	log.Printf("-----------------------------------------------------------\n")
	log.Printf("|    Mesh Version '%s' You can update Istio rule       |\n", meshVersion)
	log.Printf("-----------------------------------------------------------\n")

	return
}

// OnMeshExit cleanup proxy deployment in proxy
func (c *Connect) OnMeshExit(shadow string, clientset *kubernetes.Clientset) {
	log.Printf("Remove proxy shadow %s", shadow)
	deploymentsClient := clientset.AppsV1().Deployments(c.Namespace)
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient.Delete(shadow, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}
