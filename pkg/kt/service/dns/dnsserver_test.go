package dns

import (
	"reflect"
	"testing"
)

func Test_getDnsAddresses(t *testing.T) {
	type args struct {
		dnsOrder       []string
		upstreamDns    string
		clusterDnsPort int
	}
	tests := []struct {
		args args
		want []string
	}{
		{
			args: args{
				dnsOrder: []string{"cluster", "upstream", "tcp:upstream", "upstream:123", "tcp:upstream:123"},
				upstreamDns: "1.2.3.4",
				clusterDnsPort: 5353,
			},
			want: []string{"tcp:127.0.0.1:5353", "udp:1.2.3.4:53", "tcp:1.2.3.4:53", "udp:1.2.3.4:123", "tcp:1.2.3.4:123"},
		},
		{
			args: args{
				dnsOrder: []string{"7.8.9.0", "tcp:7.8.9.0", "7.8.9.0:123", "tcp:7.8.9.0:123"},
				upstreamDns: "1.2.3.4",
				clusterDnsPort: 5353,
			},
			want: []string{"udp:7.8.9.0:53", "tcp:7.8.9.0:53", "udp:7.8.9.0:123", "tcp:7.8.9.0:123"},
		},
		{
			args: args{
				dnsOrder: []string{"", "tcp:", ":123", "tcp:7.8.9.0:123:53"},
				upstreamDns: "1.2.3.4",
				clusterDnsPort: 5353,
			},
			want: []string{"udp::53", "tcp::53", "udp::123"},
		},
	}
	for _, tt := range tests {
		t.Run("getDnsAddresses", func(t *testing.T) {
			if got := getDnsAddresses(tt.args.dnsOrder, tt.args.upstreamDns, tt.args.clusterDnsPort); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}
