package action

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string, expose string, userHome string, pidFile string) {
	daemonRunning := util.IsDaemonRunning(pidFile)
	if !daemonRunning {
		log.Printf("'KT Connect' not runing, you can only access local app from cluster")
	} else {
		log.Printf("'KT Connect' is runing, you can access local app from cluster and localhost")
	}

	if swap == "" || expose == "" {
		err := fmt.Errorf("-replace and -expose is required")
		panic(err.Error())
	}

	factory := connect.Connect{
		Swap:       swap,
		Expose:     expose,
		Kubeconfig: action.Kubeconfig,
		Namespace:  action.Namespace,
		Image:      action.Image,
		Debug:      action.Debug,
	}

	err := factory.InitExchange()
	if err != nil {
		panic(err.Error())
	}

	defer factory.Exit()

	// SSH Remote Port forward
	factory.RemotePortForwardToPod()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	s := <-c
	log.Printf("[Exit] Signal is %s", s)
}
