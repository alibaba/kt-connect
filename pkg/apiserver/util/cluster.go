package util

import (
	ktUtil "github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// GetKubernetesClient ...
func GetKubernetesClient() (clientset kubernetes.Interface, config *restClient.Config, err error) {
	config, err = GetKubeconfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	return
}

// GetKubeconfig ...
func GetKubeconfig() (config *restClient.Config, err error) {
	kubeconfig := ktUtil.KubeConfig()
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		log.Info().Msg("Kubeconfig not found, use in-cluster mode")
		config, err := restClient.InClusterConfig()
		return config, err
	}
	log.Info().Msg("Use out-cluster config mode")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	return
}
