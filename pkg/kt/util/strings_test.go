package util

import (
	"reflect"
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
			if got := String2Map(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String2Map() = %v, want %v", got, tt.want)
			}
		})
	}
}
