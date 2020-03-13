package util

import (
	"testing"
)

func TestGetRandomSSHPort(t *testing.T) {
	type args struct {
		podIP string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "GetRandomSSHPortPodIP",
			args: args{
				podIP: "172.16.0.9",
			},
			want: "2209",
		},
		{
			name: "GetRandomSSHPortPodIP",
			args: args{
				podIP: "172.16.0.91",
			},
			want: "2291",
		},
		{
			name: "GetRandomSSHPortPodIP",
			args: args{
				podIP: "172.16.0.191",
			},
			want: "2291",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRandomSSHPort(tt.args.podIP); got != tt.want {
				t.Errorf("GetRandomSSHPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOutboundIP(t *testing.T) {
	tests := []struct {
		name        string
		wantAddress bool
	}{
		{
			name:        "shouldGetOutboundIp",
			wantAddress: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAddress := GetOutboundIP(); (gotAddress != "") != tt.wantAddress {
				t.Errorf("GetOutboundIP() = %v, want %v", gotAddress, tt.wantAddress)
			}
		})
	}
}
