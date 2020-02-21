package command

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"runtime"
	"strings"

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

func (action *Action) OpenDashboard(options *options.DaemonOptions) error {
	return nil
}

// Check check local denpendency for kt connect
func (action *Action) Check(options *options.DaemonOptions) error {
	log.Info().Msgf("system info %s-%s", runtime.GOOS, runtime.GOARCH)

	log.Info().Msg("checking ssh version")
	command := util.SSHVersion()
	err := util.BackgroundRun(command, "ssh version", true)
	if err != nil {
		log.Error().Msg("ssh is missing, please make sure command ssh is work right at your local first")
		return err
	}

	log.Info().Msg("checking kubectl version")
	command = util.KubectlVersion(options.KubeConfig)
	err = util.BackgroundRun(command, "kubectl version", true)
	if err != nil {
		log.Error().Msg("kubectl is missing, please make sure kubectl is working right at your local first")
		return err
	}

	log.Info().Msg("checking sshuttle version")
	command = util.SSHUttleVersion()
	err1 := util.BackgroundRun(command, "sshuttle version", true)
	if err1 != nil {
		log.Warn().Msg("sshuttle is missing, you can only use 'ktctl connect --method socks5' with Socks5 proxy mode")
	}

	log.Info().Msg("KT Connect is ready, enjoy it!")
	return nil
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
