package cluster

import (
	"reflect"
	"testing"

	"github.com/alibaba/kt-connect/pkg/kt/util"
	. "k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestKubernetes_CreateShadow(t *testing.T) {

	type fields struct {
	}
	type args struct {
		name      string
		namespace string
		image     string
		labels    map[string]string
		debug     bool
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		objs           []Object
		wantPodIP      string
		wantPodName    string
		wantSshcm      string
		wantCredential *util.SSHCredential
		wantErr        bool
	}{
		{
			name:   "shouldCreateShadowSuccessful",
			fields: fields{},
			args: args{
				name:      "shadow",
				namespace: "default",
				image:     "shadow/shadow",
				labels: map[string]string{
					"kt-component": "shadow-component",
					"version":      "0.0.1",
				},
				debug: true,
			},
			objs: []Object{
				buildPod(
					"shadow-pod",
					"default",
					"a",
					"172.168.1.2", map[string]string{
						"kt": "shadow",
					}),
			},
			wantPodIP:   "172.168.1.2",
			wantPodName: "shadow-pod",
			wantSshcm:   "kt-shadow-component-public-key-0.0.1",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(tt.objs...),
			}

			gotPodIP, gotPodName, gotSshcm, _, err := k.CreateShadow(tt.args.name, tt.args.namespace, tt.args.image, tt.args.labels, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.CreateShadow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPodIP != tt.wantPodIP {
				t.Errorf("Kubernetes.CreateShadow() gotPodIP = %v, want %v", gotPodIP, tt.wantPodIP)
			}
			if gotPodName != tt.wantPodName {
				t.Errorf("Kubernetes.CreateShadow() gotPodName = %v, want %v", gotPodName, tt.wantPodName)
			}
			if gotSshcm != tt.wantSshcm {
				t.Errorf("Kubernetes.CreateShadow() gotSshcm = %v, want %v", gotSshcm, tt.wantSshcm)
			}
		})
	}
}

func TestKubernetes_ClusterCrids(t *testing.T) {
	type args struct {
		podCIDR string
	}
	tests := []struct {
		name      string
		args      args
		objs      []Object
		wantCidrs []string
		wantErr   bool
	}{
		{
			name: "shouldGetClusterCrid",
			args: args{
				podCIDR: "172.168.0.0/24",
			},
			objs: []Object{
				buildNode("default", "node1", ""),
				buildPod("pod1", "default", "image", "192.168.0.7", map[string]string{
					"labe": "value",
				}),
				buildService2("default", "name", "172.168.0.18"),
			},
			wantCidrs: []string{
				"172.168.0.0/24",
				"172.168.0.0/16",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(tt.objs...),
			}
			gotCidrs, err := k.ClusterCrids(tt.args.podCIDR)
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.ClusterCrids() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCidrs, tt.wantCidrs) {
				t.Errorf("Kubernetes.ClusterCrids() = %v, want %v", gotCidrs, tt.wantCidrs)
			}
		})
	}
}
