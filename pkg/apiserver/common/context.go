package common

import (
	"github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
)

// Context Gin Global Context
type Context struct {
	Cluster cluster.Watcher
}

// Config kubernetes config
func (c *Context) Config() (config *rest.Config) {
	return c.Cluster.Config
}

// Client kubernetes rest client
func (c *Context) Client() (client kubernetes.Interface) {
	return c.Cluster.Client
}

// NamespaceLister namespace listener
func (c *Context) NamespaceLister() (lister v1.NamespaceLister) {
	return c.Cluster.NamespaceLister
}

//PodLister pod listener
func (c *Context) PodLister() (lister v1.PodLister) {
	return c.Cluster.PodLister
}

// EndpointsLister endpoint listener
func (c *Context) EndpointsLister() (lister v1.EndpointsLister) {
	return c.Cluster.EndpointsLister
}

// ServiceLister service listener
func (c *Context) ServiceLister() (lister v1.ServiceLister) {
	return c.Cluster.ServiceLister
}
