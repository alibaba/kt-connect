package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/kt/options"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(options *options.DaemonOptions) {
	if util.IsDaemonRunning(options.RuntimeOptions.PidFile) {
		err := fmt.Errorf("Connect already running %s. exit this", options.RuntimeOptions.PidFile)
		panic(err.Error())
	}

	pid := os.Getpid()
	err := ioutil.WriteFile(options.RuntimeOptions.PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		panic(err.Error())
	}

	log.Info().Msgf("Daemon Start At %d", pid)

	factory := connect.Connect{
		Kubeconfig: options.KubeConfig,
		Namespace:  options.Namespace,
		Image:      options.Image,
		Debug:      options.Debug,
		Method:     options.ConnectOptions.Method,
		ProxyPort:  options.ConnectOptions.Socke5Proxy,
		Port:       options.ConnectOptions.SSHPort,
		DisableDNS: options.ConnectOptions.DisableDNS,
		PodCIDR:    options.ConnectOptions.CIDR,
		PidFile:    options.RuntimeOptions.PidFile,
	}

	clientSet, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	options.RuntimeOptions.Shadow=workload

	labels := map[string]string{
		"kt":           workload,
		"kt-component": "connect",
		"control-by":   "kt",
	}
	for k, v := range util.String2Map(options.Labels) {
		labels[k] = v
	}

	endPointIP, podName, err := cluster.CreateShadow(
		clientSet,
		options.Namespace,
		workload,
		labels,
		options.Image,
	)

	if err != nil {
		panic(err.Error())
	}

	cidrs, err := factory.GetProxyCrids(clientSet)
	factory.StartConnect(podName, endPointIP, cidrs)

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-channel
	log.Info().Msgf("[Exit] Signal is %s", s)
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string, options *options.DaemonOptions) {
	pidFile := options.RuntimeOptions.PidFile
	daemonRunning := util.IsDaemonRunning(pidFile)
	expose := options.ExchangeOptions.Expose
	if !daemonRunning {
		log.Printf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Printf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}

	if swap == "" || expose == "" {
		err := fmt.Errorf("-expose is required")
		panic(err.Error())
	}

	factory := connect.Connect{
		Swap:       swap,
		Expose:     expose,
		Kubeconfig: options.KubeConfig,
		Namespace:  options.Namespace,
		Image:      options.Image,
		Debug:      options.Debug,
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	origin, err := clientset.AppsV1().Deployments(options.Namespace).Get(swap, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	replicas := origin.Spec.Replicas

	workload, err := factory.Exchange(options.Namespace, origin, clientset, util.String2Map(options.Labels))
	if err != nil {
		panic(err.Error())
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-c
	log.Printf("[Exit] Signal is %s", s)
	factory.HandleExchangeExit(workload, replicas, origin, clientset)
}


//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, options *options.DaemonOptions) {
	pidFile := options.RuntimeOptions.PidFile
	daemonRunning := util.IsDaemonRunning(pidFile)
	expose := options.MeshOptions.Expose

	if !daemonRunning {
		log.Printf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Printf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}

	if swap == "" || expose == "" {
		err := fmt.Errorf("-expose is required")
		panic(err.Error())
	}

	factory := connect.Connect{
		Swap:       swap,
		Expose:     expose,
		Kubeconfig: options.KubeConfig,
		Namespace:  options.Namespace,
		Image:      options.Image,
		Debug:      options.Debug,
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	workload, err := factory.Mesh(clientset, util.String2Map(options.Labels))
	if err != nil {
		panic(err.Error())
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-c
	log.Printf("[Exit] Signal is %s", s)
	factory.OnMeshExit(workload, clientset)
}
