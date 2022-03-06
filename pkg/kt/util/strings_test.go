package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestString2Map(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should handle empty string",
			args: args{
				str: "",
			},
			want: map[string]string{},
		},
		{
			name: "should covert to key value",
			args: args{
				str: "k1=v1,k2=v2",
			},
			want: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := String2Map(tt.args.str)
			require.Equal(t, got, tt.want, "String2Map() = %v, want %v", got, tt.want)
		})
	}
}

func Test_ExtractErrorMessage(t *testing.T) {
	require.Equal(t, "specfied header 'kt_version' no match mesh pod header 'vvvv'",
		ExtractErrorMessage("4:00PM ERR Update route with add failed error=\"specfied header 'kt_version' no match mesh pod header 'vvvv'\""))
	require.Empty(t, ExtractErrorMessage("4:00PM INFO Route updated."))
}
