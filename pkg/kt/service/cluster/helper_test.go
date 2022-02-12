package cluster

import (
	"k8s.io/client-go/kubernetes"
	"reflect"
	"testing"
)

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
