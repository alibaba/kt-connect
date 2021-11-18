package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString Generate random string
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// RandomDuration Generate random duration number
func RandomDuration(n int) time.Duration {
	return time.Duration(rand.Int63n(int64(n)))
}

// RandomPort Generate random number [1024, 65535)
func RandomPort() int {
	return rand.Intn(65535-1024) + 1024
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

// Append Add segment to a comma separated string
func Append(base string, inc string) string {
	if len(base) == 0 {
		return inc
	} else {
		return fmt.Sprintf("%s,%s", base, inc)
	}
}
