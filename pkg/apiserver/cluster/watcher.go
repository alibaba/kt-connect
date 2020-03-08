package cluster

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
)

var (
	errTimeout = errors.New("timed out waiting for caches to sync")
)

//Watcher Kubernetes resource watch
type Watcher struct {
	Client          kubernetes.Interface
	Config          *rest.Config
	NamespaceLister v1.NamespaceLister
	PodLister       v1.PodLister
	ServiceLister   v1.ServiceLister
	EndpointsLister v1.EndpointsLister
}

// Construct for watcher
func Construct(client kubernetes.Interface, config *rest.Config) (w Watcher, err error) {
	w = Watcher{Client: client, Config: config}

	namespaceLister, err := w.Namespaces(wait.NeverStop)
	if err != nil {
		return
	}
	w.NamespaceLister = namespaceLister

	podListener, err := w.Pods(wait.NeverStop)
	if err != nil {
		return
	}
	w.PodLister = podListener

	serviceLister, err := w.Services(wait.NeverStop)
	if err != nil {
		return
	}
	w.ServiceLister = serviceLister

	endpointLister, err := w.Endpoints(wait.NeverStop)
	if err != nil {
		return
	}
	w.EndpointsLister = endpointLister
	return
}

// ServiceListener ServiceListener
func ServiceListener(client kubernetes.Interface, stopCh <-chan struct{}) (lister v1.ServiceLister, err error) {
	w := Watcher{Client: client}
	lister, err = w.Services(stopCh)
	if err != nil {
		return
	}
	return
}

// PodListener PodListener
func PodListener(client kubernetes.Interface, stopCh <-chan struct{}) (lister v1.PodLister, err error) {
	w := Watcher{Client: client}
	lister, err = w.Pods(stopCh)
	if err != nil {
		return
	}
	return
}
