package cluster

import (
	"context"
	extV1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetAllIngressInNamespace get all ingresses in specified namespace
func (k *Kubernetes) GetAllIngressInNamespace(namespace string) (*extV1.IngressList, error) {
	return k.Clientset.ExtensionsV1beta1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{
		TimeoutSeconds: &apiTimeout,
	})
}
