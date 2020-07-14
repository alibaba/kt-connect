package cluster

import (
	"errors"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
)

var (
	errTimeout = errors.New("timed out waiting for caches to sync")
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

	namespaceLister, err := w.Namespaces(wait.NeverStop)
	if err != nil {
		return
	}
	w.NamespaceLister = namespaceLister

	podListener, err := w.Pods(wait.NeverStop)
	if err != nil {
		return
	}
	w.PodLister = podListener

	serviceLister, err := w.Services(wait.NeverStop)
	if err != nil {
		return
	}
	w.ServiceLister = serviceLister

	endpointLister, err := w.Endpoints(wait.NeverStop)
	if err != nil {
		return
	}
	w.EndpointsLister = endpointLister
	return
}

// PodListenerWithNamespace PodListener
func PodListenerWithNamespace(client kubernetes.Interface, namespace string, stopCh <-chan struct{}) (lister v1.PodLister, err error) {
	w := Watcher{Client: client}
	lister, err = w.PodsWithNamespace(namespace, stopCh)
	if err != nil {
		return
	}
	return
}

func informerFactoryWithNamespace(w *Watcher, namespace string) (factory informers.SharedInformerFactory) {
	resyncPeriod := 30 * time.Minute
	factory = informers.NewSharedInformerFactoryWithOptions(w.Client, resyncPeriod, informers.WithNamespace(namespace))
	return
}

func informerFactory(w *Watcher) (factory informers.SharedInformerFactory) {
	resyncPeriod := 30 * time.Minute
	factory = informers.NewSharedInformerFactory(w.Client, resyncPeriod)
	return
}

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

func podDeleted(obj interface{}) {
	// pod, ok := obj.(*api.Pod)
	// if ok {
	// 	fmt.Printf("Pod deleted: %s\n", pod.ObjectMeta.Name)
	// } else {
	// 	fmt.Printf("Pod deleted event: %s\n", obj)
	// }
}

// PodsWithNamespace watch pods change
func (w *Watcher) PodsWithNamespace(namespace string, stopCh <-chan struct{}) (lister v1.PodLister, err error) {
	factory := informerFactoryWithNamespace(w, namespace)
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
