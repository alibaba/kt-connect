package cluster

import (
	"github.com/alibaba/kt-connect/pkg/kt/util"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestKubernetes_CreateService(t *testing.T) {
	type args struct {
		name      string
		namespace string
		port      map[int]int
		labels    map[string]string
		annotations map[string]string
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
				port: map[int]int{8080:8080},
				labels: map[string]string{
					"label": "value",
				},
				annotations: map[string]string{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				Clientset: testclient.NewSimpleClientset(),
			}
			_, err := k.CreateService(&SvcMetaAndSpec{
				Meta: &ResourceMeta{
					Name: tt.args.name,
					Namespace: tt.args.namespace,
					Labels: map[string]string{util.ControlBy: util.KubernetesToolkit},
					Annotations: tt.args.annotations,
				},
				External: false,
				Ports: tt.args.port,
				Selectors: tt.args.labels,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Kubernetes.CreateService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
