package action

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

// Connect connect vpn to kubernetes cluster
func (action *Action) Connect(localSSHPort int, disableDNS bool, cidr string) {
	if util.IsDaemonRunning(action.PidFile) {
		err := fmt.Errorf("Connect already running. exit this")
		panic(err.Error())
	}

	factory := connect.Connect{
		Kubeconfig: action.Kubeconfig,
		Namespace:  action.Namespace,
		Name:       "kt-connect-daemon",
		Image:      action.Image,
		Port:       localSSHPort,
		DisableDNS: disableDNS,
		PodCIDR:    cidr,
		Debug:      action.Debug,
		PidFile:    action.PidFile,
	}

	err := factory.InitDaemon()
	if err != nil {
		panic(err.Error())
	}
	factory.Start()

	pid := os.Getpid()
	err = ioutil.WriteFile(action.PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Daemon Start At %d", pid)
	defer factory.Exit()

	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	s := <-channel
	log.Printf("[Exit] Signal is %s", s)
	factory.Exit()
}
