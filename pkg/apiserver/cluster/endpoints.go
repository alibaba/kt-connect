package cluster

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Endpoints informer of endpoints
func (w *Watcher) Endpoints(stopCh <-chan struct{}) (lister v1.EndpointsLister, err error) {
	factory := informerFactory(w)
	informerFactory := factory.Core().V1().Endpoints()
	informer := informerFactory.Informer()

	defer runtime.HandleCrash()

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		err = errTimeout
		runtime.HandleError(err)
		return
	}

	lister = informerFactory.Lister()
	return
}
