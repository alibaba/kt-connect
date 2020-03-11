package cluster

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func Test_getPodCirds(t *testing.T) {
	tests := []struct {
		name      string
		objs      []runtime.Object
		wantCidrs []string
		wantErr   bool
	}{
		{
			name: "should_get_pod_cird_from_pods",
			objs: []runtime.Object{
				pod("default", "a", "172.168.1.2"),
			},
			wantCidrs: []string{
				"172.168.0.0/16",
			},
			wantErr: false,
		},
		{
			name: "should_get_pod_cird_from_nodes",
			objs: []runtime.Object{
				node("default", "a", "172.168.1.0/24"),
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

			gotCidrs, err := getPodCirds(client, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("getPodCirds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCidrs, tt.wantCidrs) {
				t.Errorf("getPodCirds() = %v, want %v", gotCidrs, tt.wantCidrs)
			}
		})
	}
}

func node(namespace, name, crid string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: v1.NodeSpec{
			PodCIDR: crid,
		},
	}
}

func pod(namespace, image string, ip string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Image: image}},
		},
		Status: v1.PodStatus{
			PodIP: ip,
		},
	}
}
