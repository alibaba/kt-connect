package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

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
func (action *Action) Connect() {
	pidFile := action.Options.RuntimeOptions.PidFile
	sshPort := action.Options.ConnectOptions.SSHPort
	method := action.Options.ConnectOptions.Method
	socke5Proxy := action.Options.ConnectOptions.Socke5Proxy
	disableDNS := action.Options.ConnectOptions.DisableDNS
	cidr := action.Options.ConnectOptions.CIDR

	if util.IsDaemonRunning(pidFile) {
		err := fmt.Errorf("Connect already running %s. exit this", pidFile)
		panic(err.Error())
	}

	pid := os.Getpid()
	err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		panic(err.Error())
	}

	log.Info().Msgf("Daemon Start At %d", pid)

	factory := connect.Connect{
		Kubeconfig: action.Options.KubeConfig,
		Namespace:  action.Options.Namespace,
		Image:      action.Options.Image,
		Debug:      action.Options.Debug,
		Method:     method,
		ProxyPort:  socke5Proxy,
		Port:       sshPort,
		DisableDNS: disableDNS,
		PodCIDR:    cidr,
		PidFile:    pidFile,
	}

	clientSet, err := factory.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	labels := map[string]string{
		"kt":           workload,
		"kt-component": "connect",
		"control-by":   "kt",
	}
	for k, v := range util.String2Map(action.Options.Labels) {
		labels[k] = v
	}

	endPointIP, podName, err := factory.CreateEndpoint(
		clientSet,
		workload,
		labels,
		action.Options.Image,
		action.Options.Namespace,
	)

	if err != nil {
		panic(err.Error())
	}

	cidrs, err := factory.GetProxyCrids(clientSet)

	if err != nil {
		factory.OnConnectExit(workload, pid)
		panic(err.Error())
	}

	factory.StartConnect(podName, endPointIP, cidrs)

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-channel
	log.Info().Msgf("[Exit] Signal is %s", s)
	factory.OnConnectExit(workload, pid)
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string) {
	pidFile := action.Options.RuntimeOptions.PidFile
	daemonRunning := util.IsDaemonRunning(pidFile)
	expose := action.Options.ExchangeOptions.Expose
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
		Kubeconfig: action.Options.KubeConfig,
		Namespace:  action.Options.Namespace,
		Image:      action.Options.Image,
		Debug:      action.Options.Debug,
	}

	clientset, err := factory.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	origin, err := clientset.AppsV1().Deployments(action.Options.Namespace).Get(swap, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	replicas := origin.Spec.Replicas

	workload, err := factory.Exchange(action.Options.Namespace, origin, clientset, util.String2Map(action.Options.Labels))
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
func (action *Action) Mesh(swap string) {
	pidFile := action.Options.RuntimeOptions.PidFile
	daemonRunning := util.IsDaemonRunning(pidFile)
	expose := action.Options.MeshOptions.Expose

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
		Kubeconfig: action.Options.KubeConfig,
		Namespace:  action.Options.Namespace,
		Image:      action.Options.Image,
		Debug:      action.Options.Debug,
	}

	clientset, err := factory.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	workload, err := factory.Mesh(clientset, util.String2Map(action.Options.Labels))
	if err != nil {
		panic(err.Error())
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-c
	log.Printf("[Exit] Signal is %s", s)
	factory.OnMeshExit(workload, clientset)
}
