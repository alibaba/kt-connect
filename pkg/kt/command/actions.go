package command

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"

	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(options *options.DaemonOptions) (err error) {
	if util.IsDaemonRunning(options.RuntimeOptions.PidFile) {
		return fmt.Errorf("Connect already running %s exit this", options.RuntimeOptions.PidFile)
	}

	ch := SetUpCloseHandler(options)

	pid, err := util.WritePidFile(options.RuntimeOptions.PidFile)
	if err != nil {
		return
	}
	log.Info().Msgf("Connect Start At %d", pid)
	factory := connect.Connect{Options: options}
	clientSet, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return
	}

	if options.ConnectOptions.Dump2Hosts {
		hosts := cluster.LocalHosts(clientSet, options.Namespace)
		util.DumpHosts(hosts)
		options.ConnectOptions.Hosts = hosts
	}

	workload := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))
	options.RuntimeOptions.Shadow = workload

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
		workload,
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

	err = factory.StartConnect(podName, endPointIP, cidrs, options.Debug)
	if err != nil {
		return
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)
	return
}

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string, options *options.DaemonOptions) error {
	ch := SetUpCloseHandler(options)

	checkConnectRunning(options.RuntimeOptions.PidFile)
	expose := options.ExchangeOptions.Expose

	if swap == "" || expose == "" {
		return fmt.Errorf("-expose is required")
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return err
	}

	origin, err := clientset.AppsV1().Deployments(options.Namespace).Get(swap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	replicas := origin.Spec.Replicas

	// Prepare context inorder to remove after command exit
	options.RuntimeOptions.Origin = swap
	options.RuntimeOptions.Replicas = *replicas

	factory := connect.Connect{}
	_, err = factory.Exchange(options, origin, clientset, util.String2Map(options.Labels))
	if err != nil {
		return err
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}

//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, options *options.DaemonOptions) error {
	checkConnectRunning(options.RuntimeOptions.PidFile)

	ch := SetUpCloseHandler(options)

	expose := options.MeshOptions.Expose

	if swap == "" || expose == "" {
		return fmt.Errorf("-expose is required")
	}

	clientset, err := cluster.GetKubernetesClient(options.KubeConfig)
	if err != nil {
		return err
	}

	factory := connect.Connect{}
	_, err = factory.Mesh(swap, options, clientset, util.String2Map(options.Labels))

	if err != nil {
		return err
	}

	s := <-ch
	log.Info().Msgf("Terminal Signal is %s", s)

	return nil
}