package cluster

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"

	"reflect"
	"testing"
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
		objs           []runtime.Object
		wantPodIP      string
		wantPodName    string
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
					common.KTComponent: "shadow-component",
					common.KTVersion:   "0.0.1",
				},
				debug: true,
			},
			objs: []runtime.Object{
				buildPod(
					"shadow-pod",
					"default",
					"a",
					"172.168.1.2", map[string]string{
						"kt-name": "shadow",
					}),
			},
			wantPodIP:   "172.168.1.2",
			wantPodName: "shadow-pod",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(tt.objs...),
			}

			envs := make(map[string]string)
			annotations := make(map[string]string)
			option := options.DaemonOptions{Namespace: tt.args.namespace, Image: tt.args.image, Debug: tt.args.debug}
			gotPodIP, gotPodName, gotSshcm, _, err := k.GetOrCreateShadow(tt.args.name, &option, tt.args.labels, annotations, envs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.GetOrCreateShadow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPodIP != tt.wantPodIP {
				t.Errorf("Kubernetes.GetOrCreateShadow() gotPodIP = %v, want %v", gotPodIP, tt.wantPodIP)
			}
			if gotPodName != tt.wantPodName {
				t.Errorf("Kubernetes.GetOrCreateShadow() gotPodName = %v, want %v", gotPodName, tt.wantPodName)
			}
			if gotSshcm == "" {
				t.Errorf("Kubernetes.GetOrCreateShadow() gotSshcm = %v", gotSshcm)
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
		objs      []runtime.Object
		wantCidrs []string
		wantErr   bool
	}{
		{
			name: "shouldGetClusterCrid",
			args: args{
				podCIDR: "172.168.0.0/24",
			},
			objs: []runtime.Object{
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
			ops := &options.ConnectOptions{
				CIDR: tt.args.podCIDR,
			}
			gotCidrs, err := k.ClusterCrids("default", ops)
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

func TestKubernetes_CreateService(t *testing.T) {
	type args struct {
		name      string
		namespace string
		port      int
		labels    map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "shouldCreateService",
			args: args{
				name:      "svc-name",
				namespace: "default",
				port:      8080,
				labels: map[string]string{
					"label": "value",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(),
			}
			_, err := k.CreateService(tt.args.name, tt.args.namespace, false, tt.args.port, tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.CreateService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestKubernetes_ScaleTo(t *testing.T) {
	type args struct {
		deployment string
		namespace  string
		replicas   int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		objs    []runtime.Object
	}{
		{
			name: "shouldScaleDeployToReplicas",
			args: args{
				deployment: "app",
				namespace:  "default",
				replicas:   int32(2),
			},
			objs: []runtime.Object{
				&appv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app",
						Namespace: "default",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(tt.objs...),
			}
			if err := k.ScaleTo(tt.args.deployment, tt.args.namespace, &tt.args.replicas); (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.ScaleTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
