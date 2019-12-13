package util

import (
	"math/rand"
	"strings"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString Generate RandomString
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// String2Map Convert parameter string to real map "k1=v1,k2=v2" -> {"k1":"v1","k2","v2"}
func String2Map(str string) map[string]string {
	res := make(map[string]string)
	splitStr := strings.Split(str, ",")
	for _, item := range splitStr {
		index := strings.Index(item, "=")
		if index > 0 {
			res[item[0:index]] = item[index+1:]
		}
	}
	return res
}