package cluster

import (
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestKubernetes_ClusterCidrs(t *testing.T) {
	type args struct {
		IncludeIps []string
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
				IncludeIps: []string{
					"10.10.10.0/24",
				},
			},
			objs: []runtime.Object{
				buildPod("default", "pod1", "image", "172.168.0.7", map[string]string{"label": "value"}),
				buildPod("default", "pod2", "image", "172.168.0.8", map[string]string{"label": "value"}),
				buildPod("default", "pod3", "image", "172.167.0.7", map[string]string{"label": "value"}),
				buildPod("default", "pod4", "image", "172.167.0.8", map[string]string{"label": "value"}),
				buildService("default", "svc1", "192.168.0.18"),
				buildService("default", "svc2", "192.168.1.18"),
			},
			wantCidrs: []string{
				"192.168.0.0/16",
				"172.168.0.0/24",
				"172.167.0.0/24",
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
			opt.Get().ConnectOptions.IncludeIps = strings.Join(tt.args.IncludeIps, ",")
			opt.Get().RuntimeStore.RestConfig = &rest.Config{ Host: "" }
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

func Test_calculateMinimalIpRange(t *testing.T) {
	tests := []struct {
		name string
		ips []string
		miniRange []string
	}{
		{
			name: "1 range",
			ips: []string{"1.2.3.4", "1.2.3.100"},
			miniRange: []string{"1.2.3.0/24"},
		},
		{
			name: "2 ranges",
			ips: []string{"1.2.3.4", "2.3.4.5", "1.2.3.100", "2.3.5.5"},
			miniRange: []string{"1.2.3.0/24", "2.3.0.0/16"},
		},
		{
			name: "duplicate address",
			ips: []string{"1.2.3.4", "1.2.3.4", "1.2.3.4", "1.2.3.4"},
			miniRange: []string{"1.2.3.4/32"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			realRange := calculateMinimalIpRange(test.ips)
			if !util.ArrayEquals(test.miniRange, realRange) {
				t.Fatalf("range %v not as expect as %v", realRange, test.miniRange)
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

func Test_removeCidrOf(t *testing.T) {
	cidrs := []string{"192.168.10.0/24", "192.168.11.0/24", "192.168.12.0/24"}
	expectedCidrs := []string{"192.168.10.0/24", "192.168.12.0/24"}
	realCidrs := removeCidrOf(cidrs, "192.168.11.1/26")
	require.Equal(t, len(expectedCidrs), len(realCidrs))
	for i := 0; i < len(expectedCidrs); i++ {
		require.Equal(t, expectedCidrs[i], realCidrs[i])
	}
}

func Test_isPartOfRange(t *testing.T) {
	tests := []struct {
		range1 string
		range2 string
		res bool
	}{
		{range1: "192.168.10.11/32", range2: "192.168.10.11/32", res: true},
		{range1: "192.168.10.0/24", range2: "192.168.10.11/32", res: true},
		{range1: "192.168.11.0/24", range2: "192.168.10.11/32", res: false},
	}
	for _, test := range tests {
		t.Run(test.range2, func(t *testing.T) {
			require.Equal(t, test.res, isPartOfRange(test.range1, test.range2))
		})
	}
}

func Test_ipRangeToBin(t *testing.T) {
	ipNum, _ := ipRangeToBin("100.25.255.0/24")
	expIp := []int{0, 1, 1, 0, 0, 1, 0, 0,
		0, 0, 0, 1, 1, 0, 0, 1,
		1, 1, 1, 1, 1, 1, 1, 1,
		-1, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < 32; i++ {
		require.Equal(t, expIp[i], ipNum[i], "failed on index: %d", i)
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
	require.Equal(t, "100.25.255.0/32", binToIpRange(ipBin, false))

	ipBin = [32]int{0, 1, 1, 0, 0, 1, 0, 0,
		0, 0, 0, 1, 1, 0, 0, 1,
		1, 1, 1, 1, 1, 1, 1, -1,
		0, 0, 0, 0, 0, 0, 0, 0}
	require.Equal(t, "100.25.254.0/23", binToIpRange(ipBin, false))
}

func buildService(namespace, name, clusterIP string) *coreV1.Service {
	return &coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: coreV1.ServiceSpec{
			ClusterIP: clusterIP,
		},
	}
}

func buildPod(namespace, name, image string, ip string, labels map[string]string) *coreV1.Pod {
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
