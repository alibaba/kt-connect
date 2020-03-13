package cluster

import (
	"testing"

	. "k8s.io/apimachinery/pkg/runtime"

	"github.com/alibaba/kt-connect/pkg/kt/util"
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
				pod(
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
