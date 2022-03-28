package cluster

import (
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

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
					Spec: appv1.DeploymentSpec {
						Replicas: func() *int32 {
							i := int32(0)
							return &i
						}(),
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
