package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
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
					common.KtComponent: "shadow-component",
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
			option := options.DaemonOptions{Namespace: tt.args.namespace, Image: tt.args.image, Debug: tt.args.debug,
				ConnectOptions: &options.ConnectOptions{ShareShadow: false}}
			gotPodIP, gotPodName, gotSshcm, _, err := GetOrCreateShadow(context.TODO(), k, tt.args.name, &option, tt.args.labels, annotations, envs)
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