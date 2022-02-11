package cluster

import (
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"strings"
	"testing"
)

func TestKubernetes_ClusterCidrs(t *testing.T) {
	type args struct {
		extraIps []string
	}
	tests := []struct {
		name      string
		args      args
		objs      []runtime.Object
		wantCidrs []string
		wantErr   bool
	}{
		{
			name: "shouldGetClusterCidr",
			args: args{
				extraIps: []string{
					"10.10.10.0/24",
				},
			},
			objs: []runtime.Object{
				buildNode("default", "node1", "192.168.0.0/24"),
				buildPod("default", "pod1", "image", "192.168.0.7", map[string]string{"labe": "value"}),
				buildService("default", "svc1", "172.168.0.18"),
				buildService("default", "svc2", "172.168.1.18"),
			},
			wantCidrs: []string{
				"192.168.0.0/24",
				"172.168.0.0/16",
				"10.10.10.0/24",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(tt.objs...),
			}
			opt.Get().ConnectOptions.IncludeIps = strings.Join(tt.args.extraIps, ",")
			gotCidrs, err := k.ClusterCidrs("default")
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.ClusterCidrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCidrs, tt.wantCidrs) {
				t.Errorf("Kubernetes.ClusterCidrs() = %v, want %v", gotCidrs, tt.wantCidrs)
			}
		})
	}
}
