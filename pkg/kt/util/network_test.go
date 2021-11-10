package util

import (
	"testing"
)

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
