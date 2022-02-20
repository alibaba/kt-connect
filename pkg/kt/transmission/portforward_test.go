package transmission

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_parseReqHost(t *testing.T) {
	type args struct {
		host string
		path string
	}
	tests := []struct {
		name string
		args args
		want *url.URL
	}{
		{
			name: "without port",
			args: args{
				host: "http://1.2.3.4",
				path: "/api/v1/test",
			},
			want: &url.URL{Scheme: "http", Host: "1.2.3.4", Path: "/api/v1/test"},
		},
		{
			name: "with port",
			args: args{
				host: "https://1.2.3.4:6443",
				path: "/api/v1/test",
			},
			want: &url.URL{Scheme: "https", Host: "1.2.3.4:6443", Path: "/api/v1/test"},
		},
		{
			name: "with extra url",
			args: args{
				host: "https://1.2.3.4:6443/k8s/cluster",
				path: "/api/v1/test",
			},
			want: &url.URL{Scheme: "https", Host: "1.2.3.4:6443", Path: "/k8s/cluster/api/v1/test"},
		},
		{
			name: "with extra url and slash ending",
			args: args{
				host: "https://1.2.3.4:6443/k8s/cluster/",
				path: "/api/v1/test",
			},
			want: &url.URL{Scheme: "https", Host: "1.2.3.4:6443", Path: "/k8s/cluster/api/v1/test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := parseReqHost(tt.args.host, tt.args.path); !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("parseReqHost() = %v, want %v", *got, *tt.want)
			}
		})
	}
}
