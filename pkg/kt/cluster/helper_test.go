package cluster

import (
	"context"
	"github.com/stretchr/testify/require"
	"reflect"
	"strconv"
	"testing"

	coreV1 "k8s.io/api/core/v1"
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
			gotCidrs, err := getPodCidrs(context.TODO(), client, "default", "")
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
	tests := []struct {
		name     string
		objs      []runtime.Object
		wantCidr []string
		wantErr  bool
	}{
		{
			name: "should_get_service_cidr_by_svc_sample",
			objs: []runtime.Object{
				buildPod("POD1", "default", "a", "172.168.1.2", map[string]string{}),
			},
			wantErr:  false,
			wantCidr: []string{"173.168.0.0/16"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testclient.NewSimpleClientset(tt.objs...)
			gotCidr, err := getServiceCidr(context.TODO(), client, "default")
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

func Test_calculateMinimalIpRange(t *testing.T) {
	tests := []struct {
		name string
		ips []string
		miniRange []string
	}{
		{
			name: "1 range",
			ips: []string{"1.2.3.4", "1.2.3.100"},
			miniRange: []string{"1.2.3.0/25"},
		},
		{
			name: "2 ranges",
			ips: []string{"1.2.3.4", "2.3.4.5", "1.2.3.100", "2.3.5.5"},
			miniRange: []string{"1.2.3.0/25", "2.3.0.0/23"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			realRange := calculateMinimalIpRange(test.ips)
			require.Equal(t, len(test.miniRange), len(realRange), "range length should equal for %s", test.name)
			for i := 0; i < len(realRange); i++ {
				found := false
				for j := 0; j < len(test.miniRange); j++ {
					if realRange[i] == test.miniRange[j] {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("range %s not found for %s", realRange[i], test.name)
				}
			}
		})
	}
}

func Test_decToBin(t *testing.T) {
	tests := []struct {
		num int
		bin []int
	}{
		{num: 0, bin: []int{0, 0, 0, 0, 0, 0, 0, 0}},
		{num: 25, bin: []int{0, 0, 0, 1, 1, 0, 0, 1}},
		{num: 100, bin: []int{0, 1, 1, 0, 0, 1, 0, 0}},
		{num: 255, bin: []int{1, 1, 1, 1, 1, 1, 1, 1}},
	}
	for _, test := range tests {
		t.Run(strconv.Itoa(test.num), func(t *testing.T) {
			res := decToBin(test.num)
			require.Equal(t, len(res), len(test.bin))
			for i := 0; i < len(res); i++ {
				require.Equal(t, res[i], test.bin[i])
			}
		})
	}
}

func Test_ipToBin(t *testing.T) {
	ipNum, _ := ipToBin("100.25.255.0")
	expIp := []int{0, 1, 1, 0, 0, 1, 0, 0,
		0, 0, 0, 1, 1, 0, 0, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
		0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < 32; i++ {
		require.Equal(t, expIp[i], ipNum[i], "failed on index: %d", i)
	}
}

func Test_binToIpRange(t *testing.T)  {
	ipBin := [32]int{0, 1, 1, 0, 0, 1, 0, 0,
		0, 0, 0, 1, 1, 0, 0, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
		0, 0, 0, 0, 0, 0, 0, 0}
	require.Equal(t, "100.25.255.0/32", binToIpRange(ipBin))

	ipBin = [32]int{0, 1, 1, 0, 0, 1, 0, 0,
		0, 0, 0, 1, 1, 0, 0, 1,
		1, 1, 1, 1, 1, 1, 1, -1,
		0, 0, 0, 0, 0, 0, 0, 0}
	require.Equal(t, "100.25.254.0/23", binToIpRange(ipBin))
}

func buildService(namespace, name, clusterIP string) *coreV1.Service {
	return &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: coreV1.ServiceSpec{
			ClusterIP: clusterIP,
		},
	}
}

func buildNode(namespace, name, cidr string) *coreV1.Node {
	return &coreV1.Node{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: coreV1.NodeSpec{
			PodCIDR: cidr,
		},
	}
}

func buildPod(name, namespace, image string, ip string, labels map[string]string) *coreV1.Pod {
	return &coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: coreV1.PodSpec{
			Containers: []coreV1.Container{{Image: image}},
		},
		Status: coreV1.PodStatus{
			PodIP: ip,
			Phase: coreV1.PodRunning,
		},
	}
}
