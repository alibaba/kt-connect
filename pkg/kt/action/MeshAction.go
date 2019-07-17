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

//Mesh exchange kubernetes workload
func (action *Action) Mesh(swap string, expose string, userHome string, pidFile string) {
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

	clientset, err := factory.GetClientSet()
	if err != nil {
		panic(err.Error())
	}

	shadow, err := factory.Mesh(clientset)
	if err != nil {
		panic(err.Error())
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	s := <-c
	log.Printf("[Exit] Signal is %s", s)
	factory.OnMeshExit(shadow, clientset)
}
