package birdseye

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSortServiceArray(t *testing.T) {
	type args struct {
		svc       [][]string
		compIndex int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "sort by 0",
			args: args{
				svc: [][]string{
					{"ABCD", "3"},
					{"B", "6"},
					{"ABD", "5"},
					{"AB", "1"},
					{"ABC", "2"},
					{"ABCE", "4"},
				},
				compIndex: 0,
			},
		},
		{
			name: "sort by 1",
			args: args{
				svc: [][]string{
					{"6", "B"},
					{"4", "ABCE"},
					{"5", "ABD"},
					{"3", "ABCD"},
					{"2", "ABC"},
					{"1", "AB"},
				},
				compIndex: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortServiceArray(tt.args.svc, tt.args.compIndex)
			sortIndex := 0
			if tt.args.compIndex == 0 {
				sortIndex = 1
			}
			for i := 0; i < len(tt.args.svc); i++ {
				require.Equal(t, fmt.Sprintf("%d", i + 1), tt.args.svc[i][sortIndex])
			}
		})
	}
}
