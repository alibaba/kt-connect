package cluster

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Services informer of service
func (w *Watcher) Services(stopCh <-chan struct{}) (lister v1.ServiceLister, err error) {
	factory := informerFactory(w)
	serviceformer := factory.Core().V1().Services()
	informer := serviceformer.Informer()

	defer runtime.HandleCrash()

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		err = errTimeout
		runtime.HandleError(err)
		return
	}

	lister = serviceformer.Lister()
	return
}
