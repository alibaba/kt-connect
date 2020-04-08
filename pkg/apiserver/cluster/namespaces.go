package cluster

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Namespaces informer of namespace
func (w *Watcher) Namespaces(stopCh <-chan struct{}) (lister v1.NamespaceLister, err error) {
	factory := informerFactory(w)
	informerFactory := factory.Core().V1().Namespaces()
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
