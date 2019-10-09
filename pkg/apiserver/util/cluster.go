package util

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubernetesClient() (clientset kubernetes.Interface, config *restclient.Config, err error) {
	config, err = GetKubeconfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	return
}

func GetKubeconfig() (config *restclient.Config, err error) {
	kubeconfig := filepath.Join(homeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		log.Printf("kubeconfig not found, use InCluster Mode")
		config, err := restclient.InClusterConfig()
		return config, err
	}
	log.Printf("Use OutCluster Config Mode")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	return
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
