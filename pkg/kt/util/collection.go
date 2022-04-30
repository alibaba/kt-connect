package util

import (
	"reflect"
)

// Contains check whether obj exist in target, the type of target can be an array, slice or map
func Contains(obj any, target any) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

func MapContains(subset, fullset map[string]string) bool {
	for sk, sv := range subset {
		if fullset[sk] != sv {
			return false
		}
	}
	return true
}

func MapEquals(src, target map[string]string) bool {
	return len(src) == len(target) && MapContains(src, target)
}

func MapPut(m map[string]string, k, v string) map[string]string {
	if m == nil {
		return map[string]string{k: v}
	}
	m[k] = v
	return m
}

func MergeMap(m1, m2 map[string]string) map[string]string {
	cp := make(map[string]string)
	if m1 != nil {
		for key, value := range m1 {
			cp[key] = value
		}
	}
	if m2 != nil {
		for key, value := range m2 {
			cp[key] = value
		}
	}
	return cp
}

func ArrayEquals(src, target []string) bool {
	if len(src) != len(target) {
		return false
	}
	for i := 0; i < len(src); i++ {
		found := false
		for j := 0; j < len(target); j++ {
			if src[i] == target[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func ArrayDelete(arr []string, item string) []string {
	count := 0
	for _, v := range arr {
		if v == item {
			count++
		}
	}
	if count == 0 {
		return arr
	}
	newArr := make([]string, len(arr) - count)
	i := 0
	for _, v := range arr {
		if v != item {
			newArr[i] = v
			i++
		}
	}
	return newArr
}
