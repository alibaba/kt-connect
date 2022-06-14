package birdseye

import "strings"

func SortServiceArray(svc [][]string, compIndex int) {
	if len(svc) == 0 {
		return
	}
	swapper := make([]string, len(svc[0]))
	for i := 0; i < len(svc) - 1; i++ {
		flag := false
		for j := 0; j < len(svc) - 1 - i; j++ {
			if strings.Compare(svc[j][compIndex], svc[j + 1][compIndex]) > 0 {
				for k := 0; k < len(swapper); k++ {
					swapper[k] = svc[j][k]
					svc[j][k] = svc[j + 1][k]
					svc[j + 1][k] = swapper[k]
				}
				flag = true
			}
		}
		if !flag {
			break
		}
	}
}
