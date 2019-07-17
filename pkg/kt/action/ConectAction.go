package action

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(localSSHPort int, disableDNS bool, cidr string) {
	if util.IsDaemonRunning(action.PidFile) {
		err := fmt.Errorf("Connect already running %s. exit this", action.PidFile)
		panic(err.Error())
	}

	pid := os.Getpid()
	err := ioutil.WriteFile(action.PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Daemon Start At %d", pid)

	factory := connect.Connect{
		Kubeconfig: action.Kubeconfig,
		Namespace:  action.Namespace,
		Image:      action.Image,
		Port:       localSSHPort,
		DisableDNS: disableDNS,
		PodCIDR:    cidr,
		Debug:      action.Debug,
		PidFile:    action.PidFile,
	}

	clientSet, err := factory.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	endpointName := fmt.Sprintf("kt-connect-daemon-%s", strings.ToLower(util.RandomString(5)))

	endPointIP, err := factory.CreateEndpoint(
		clientSet,
		endpointName,
		map[string]string{
			"kt":           endpointName,
			"kt-component": "connect",
			"control-by":   "kt",
		},
		action.Image,
		action.Namespace,
	)

	defer factory.OnConnectExit(endpointName, pid)

	if err != nil {
		panic(err.Error())
	}

	cidrs, err := factory.GetProxyCrids(clientSet)

	if err != nil {
		panic(err.Error())
	}

	factory.StartVPN(endpointName, endPointIP, cidrs)

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	s := <-channel
	log.Printf("[Exit] Signal is %s", s)
	factory.OnConnectExit(endpointName, pid)
}
