package connect

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

// Exchange exchange request to local
func Exchange(options *options.DaemonOptions, origin *v1.Deployment, clientset *kubernetes.Clientset) (workload string, err error) {
	workload = origin.GetObjectMeta().GetName() + "-kt-" + strings.ToLower(util.RandomString(5))
	options.RuntimeOptions.Shadow = workload
	podIP, podName, err := createExchangeShadow(origin, workload, clientset, options)
	down := int32(0)
	scaleTo(origin, options.Namespace, clientset, &down)
	remotePortForward(options.ExchangeOptions.Expose, options.KubeConfig, options.Namespace, podName, podIP, options.Debug)
	return
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

func createExchangeShadow(
	origin *v1.Deployment,
	workload string,
	clientset *kubernetes.Clientset,
	options *options.DaemonOptions,
) (podIP string, podName string, err error) {
	namespace := options.Namespace
	image := options.Image
	log.Info().Msgf("Create Exchange shadow %s in namespace %s", workload, namespace)
	labels := util.Labels(workload, "exchange", origin.Spec.Selector.MatchLabels, options.Labels)
	podIP, podName, err = cluster.CreateShadow(clientset, workload, labels, namespace, image)
	return
}
