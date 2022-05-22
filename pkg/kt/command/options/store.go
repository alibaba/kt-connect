package options

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var Store = &RuntimeStore{}

// RuntimeStore ...
type RuntimeStore struct {
	// Clientset for kubernetes operation
	Clientset kubernetes.Interface
	// RestConfig kubectl config
	RestConfig *rest.Config
	// Version ktctl version
	Version string
	// Component current sub-command (connect, exchange, mesh or preview)
	Component string
	// Shadow pod name
	Shadow string
	// Router pod name
	Router string
	// Mesh version of mesh pod
	Mesh string
	// Origin the origin deployment or service name
	Origin string
	// Replicas the origin replicas
	Replicas int32
	// Service exposed service name
	Service string
}
