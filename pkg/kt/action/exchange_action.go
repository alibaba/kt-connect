package action

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alibaba/kt-connect/pkg/kt/connect"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Exchange exchange kubernetes workload
func (action *Action) Exchange(swap string, expose string) {
	pidFile := action.Options.RuntimeOptions.PidFile
	daemonRunning := util.IsDaemonRunning(pidFile)
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
