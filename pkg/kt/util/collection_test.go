package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContains(t *testing.T) {
	type args struct {
		obj    any
		target any
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
			require.Equal(t, Contains(tt.args.target, tt.args.obj), tt.want)
		})
	}
}

func TestMapContains(t *testing.T) {
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

func TestMapPut(t *testing.T) {
	type obj struct {
		label map[string]string
	}
	type args struct {
		origin obj
		key string
		value string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normal map",
			args: args {
				origin: obj { label: map[string]string{"a": "b"} },
				key: "c",
				value: "d",
			},
			want: map[string]string {"a": "b", "c": "d"},
		},
		{
			name: "nil map",
			args: args {
				origin: obj { label: nil },
				key: "c",
				value: "d",
			},
			want: map[string]string {"c": "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := MapPut(tt.args.origin.label, tt.args.key, tt.args.value)
			require.True(t, MapEquals(target, tt.want), "%v != %v", target, tt.want)
		})
	}
}

func TestArrayDelete(t *testing.T) {
	tests := []struct {
		name string
		origin []string
		target string
		want   []string
	}{
		{
			name: "not exist",
			origin: []string{"a", "b"},
			target: "c",
			want: []string{"a", "b"},
		},
		{
			name: "exist at head",
			origin: []string{"a", "b"},
			target: "a",
			want: []string{"b"},
		},
		{
			name: "exist at end",
			origin: []string{"a", "b"},
			target: "b",
			want: []string{"a"},
		},
		{
			name: "exist multiple",
			origin: []string{"a", "b", "a", "c", "a"},
			target: "a",
			want: []string{"b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, ArrayDelete(tt.origin, tt.target), tt.want)
		})
	}
}