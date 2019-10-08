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

func (c *Context) Config() (config *rest.Config) {
	return c.Cluster.Config
}

func (c *Context) Client() (client kubernetes.Interface) {
	return c.Cluster.Client
}

func (c *Context) NamespaceLister() (lister v1.NamespaceLister) {
	return c.Cluster.NamespaceLister
}

func (c *Context) PodLister() (lister v1.PodLister) {
	return c.Cluster.PodLister
}

func (c *Context) EndpointsLister() (lister v1.EndpointsLister) {
	return c.Cluster.EndpointsLister
}

func (c *Context) ServiceLister() (lister v1.ServiceLister) {
	return c.Cluster.ServiceLister
}
