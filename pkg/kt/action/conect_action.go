package action

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
)

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(sshPort int, method string, socke5Proxy int, disableDNS bool, cidr string) {
	pidFile := action.Options.RuntimeOptions.PidFile

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
		Debug:      action.Debug,
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
