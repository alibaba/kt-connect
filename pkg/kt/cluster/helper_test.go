package cluster

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func Test_getPodCidrs(t *testing.T) {
	tests := []struct {
		name      string
		objs      []runtime.Object
		wantCidrs []string
		wantErr   bool
	}{
		{
			name: "should_get_pod_cidr_from_pods",
			objs: []runtime.Object{
				buildPod("POD1", "default", "a", "172.168.1.2", map[string]string{}),
			},
			wantCidrs: []string{
				"172.168.0.0/16",
			},
			wantErr: false,
		},
		{
			name: "should_get_pod_cidr_from_nodes",
			objs: []runtime.Object{
				buildNode("default", "a", "172.168.1.0/24"),
			},
			wantCidrs: []string{
				"172.168.1.0/24",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := testclient.NewSimpleClientset(tt.objs...)

			gotCidrs, err := getPodCidrs(context.TODO(), client, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("getPodCidrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCidrs, tt.wantCidrs) {
				t.Errorf("getPodCidrs() = %v, want %v", gotCidrs, tt.wantCidrs)
			}
		})
	}
}

func Test_getServiceCidr(t *testing.T) {
	type args struct {
		serviceList []v1.Service
	}
	tests := []struct {
		name     string
		args     args
		wantCidr []string
		wantErr  bool
	}{
		{
			name: "should_get_service_cidr_by_svc_sample",
			args: args{
				[]v1.Service{
					buildService("default", "name", "173.168.0.1"),
				},
			},
			wantErr:  false,
			wantCidr: []string{"173.168.0.0/16"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCidr, err := getServiceCidr(tt.args.serviceList)
			if (err != nil) != tt.wantErr {
				t.Errorf("getServiceCidr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCidr, tt.wantCidr) {
				t.Errorf("getServiceCidr() = %v, want %v", gotCidr, tt.wantCidr)
			}
		})
	}
}

func Test_getKubernetesClient(t *testing.T) {
	type args struct {
		kubeConfig string
	}
	tests := []struct {
		name          string
		args          args
		wantClientset *kubernetes.Clientset
		wantErr       bool
	}{
		{
			name: "shouldFailGetKubernetesClientWhenKubeConfigIsEmpty",
			args: args{
				kubeConfig: "",
			},
			wantClientset: nil,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClientset, err := getKubernetesClient(tt.args.kubeConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKubernetesClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClientset, tt.wantClientset) {
				t.Errorf("getKubernetesClient() = %v, want %v", gotClientset, tt.wantClientset)
			}
		})
	}
}

func buildService(namespace, name, clusterIP string) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1.ServiceSpec{
			ClusterIP: clusterIP,
		},
	}
}

func buildService2(namespace, name, clusterIP string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1.ServiceSpec{
			ClusterIP: clusterIP,
		},
	}
}

func buildNode(namespace, name, cidr string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1.NodeSpec{
			PodCIDR: cidr,
		},
	}
}

func buildPod(name, namespace, image string, ip string, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Image: image}},
		},
		Status: v1.PodStatus{
			PodIP: ip,
			Phase: v1.PodRunning,
		},
	}
}
