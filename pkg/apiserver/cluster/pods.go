package cluster

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func podDeleted(obj interface{}) {
	// pod, ok := obj.(*api.Pod)
	// if ok {
	// 	fmt.Printf("Pod deleted: %s\n", pod.ObjectMeta.Name)
	// } else {
	// 	fmt.Printf("Pod deleted event: %s\n", obj)
	// }
}

// Pods watch pods change
func (w *Watcher) Pods(stopCh <-chan struct{}) (lister v1.PodLister, err error) {
	factory := informerFactory(w)
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()

	defer runtime.HandleCrash()

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		err = errTimeout
		runtime.HandleError(err)
		return
	}

	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			// AddFunc:    podCreated,
			DeleteFunc: podDeleted,
		},
	)

	lister = podInformer.Lister()
	return
}
