package cluster

import (
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Namespaces informer of namespace
func (w *Watcher) Namespaces() (lister v1.NamespaceLister, err error) {
	resyncPeriod := 30 * time.Minute
	stopCh := wait.NeverStop
	factory := informers.NewSharedInformerFactory(w.Client, resyncPeriod)
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
