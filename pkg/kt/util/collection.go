package util

import "reflect"

// Contains check whether obj exist in target, the type of target can be array, slice or map
func Contains(obj interface{}, target interface{}) bool {
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
		find := false
		for fk, fv := range fullset {
			if sk == fk && sv == fv {
				find = true
				break
			}
		}
		if !find {
			return false
		}
	}
	return true
}
