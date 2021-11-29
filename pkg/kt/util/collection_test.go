package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContains(t *testing.T) {
	type args struct {
		obj    interface{}
		target interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "list contains",
			args: args{
				obj: "b",
				target: []string {"a", "b", "c"},
			},
			want: true,
		},
		{
			name: "list not contains",
			args: args{
				obj: "d",
				target: []string {"a", "b", "c"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, Contains(tt.args.obj, tt.args.target), tt.want)
		})
	}
}

func TestMapEquals(t *testing.T) {
	type args struct {
		subset map[string]string
		fullset map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal contains",
			args: args{
				subset: map[string]string {"a": "b", "c": "d"},
				fullset: map[string]string {"c": "d", "a": "b"},
			},
			want: true,
		},
		{
			name: "overlap contains",
			args: args{
				subset: map[string]string {"a": "b", "c": "d"},
				fullset: map[string]string {"c": "d", "e": "f", "a": "b"},
			},
			want: true,
		},
		{
			name: "not contains",
			args: args{
				subset: map[string]string {"a": "b", "e": "f", "c": "d"},
				fullset: map[string]string {"c": "d", "a": "b"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, MapContains(tt.args.subset, tt.args.fullset), tt.want)
		})
	}
}
