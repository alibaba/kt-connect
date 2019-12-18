package command

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Action cmd action
type Action struct {
	Options *options.DaemonOptions
}

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(options *options.DaemonOptions) (err error) {
	if options.ConnectOptions.Method != "socks5" {
		checkAndWritePidFile(options.RuntimeOptions.PidFile)
	}

	clientSet, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return
	}

	options.RuntimeOptions.Shadow = fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	labels := util.Labels(options.RuntimeOptions.Shadow, "connect", map[string]string{}, options.Labels)

	endPointIP, podName, err := cluster.CreateShadow(
		clientSet,
		options.RuntimeOptions.Shadow,
		labels,
		options.Namespace,
		options.Image,
	)

	if err != nil {
		return
	}

	cidrs, err := util.GetCirds(clientSet, options.ConnectOptions.CIDR)
	if err != nil {
		return
	}

	connect.StartConnect(podName, endPointIP, cidrs, options)
	return
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string, options *options.DaemonOptions) {
	checkConnectRunning(options.RuntimeOptions.PidFile)
	expose := options.ExchangeOptions.Expose

	if swap == "" || expose == "" {
		err := fmt.Errorf("-expose is required")
		panic(err.Error())
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

	// Prepare context inorder to remove after command exit
	options.RuntimeOptions.Origin = swap
	options.RuntimeOptions.Replicas = *replicas

	_, err = connect.Exchange(options, origin, clientset)
	if err != nil {
		panic(err.Error())
	}
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, options *options.DaemonOptions) {
	checkConnectRunning(options.RuntimeOptions.PidFile)
	expose := options.MeshOptions.Expose

	if swap == "" || expose == "" {
		err := fmt.Errorf("-expose is required")
		panic(err.Error())
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	_, err = connect.Mesh(swap, options, clientset)
	if err != nil {
		panic(err.Error())
	}
}

// checkAndWritePid check PidFile present and write current pid
func checkAndWritePidFile(pidFile string) (err error) {
	if util.IsDaemonRunning(pidFile) {
		err = fmt.Errorf("Connect already running %s exit this", pidFile)
		panic(err)
	}
	pid, err := util.WritePidFile(pidFile)
	log.Info().Msgf("Connect Start At %d", pid)
	return
}

// checkConnectRunning check connect is running and print help msg
func checkConnectRunning(pidFile string) {
	daemonRunning := util.IsDaemonRunning(pidFile)
	if !daemonRunning {
		log.Info().Msgf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Info().Msgf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}
}
