package cluster

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Endpoints informor
func (w *Watcher) Endpoints() (lister v1.EndpointsLister, err error) {

	resyncPeriod := 30 * time.Minute
	stopCh := wait.NeverStop
	factory := informers.NewSharedInformerFactory(w.Client, resyncPeriod)
	informerFactory := factory.Core().V1().Endpoints()
	informer := informerFactory.Informer()

	defer runtime.HandleCrash()

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	lister = informerFactory.Lister()
	return
}
