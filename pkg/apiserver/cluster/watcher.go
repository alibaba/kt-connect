package cluster

import (
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
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
	w.Client = client

	namespaceLister, err := w.Namespaces()
	if err != nil {
		return
	}
	w.NamespaceLister = namespaceLister

	podListener, err := w.Pods()
	if err != nil {
		return
	}
	w.PodLister = podListener

	serviceLister, err := w.Services()
	if err != nil {
		return
	}
	w.ServiceLister = serviceLister

	endpointLister, err := w.Endpoints()
	if err != nil {
		return
	}
	w.EndpointsLister = endpointLister
	return
}
