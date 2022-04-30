package util

import (
	"fmt"
	"math/rand"
	"regexp"
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

// RandomSeconds Generate random duration number in second
func RandomSeconds(min, max int) time.Duration {
	val := rand.Intn(max)
	if val < min {
		val = min
	}
	return time.Duration(val) * time.Second
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

// RemoveColor remove shell color character in text
func RemoveColor(msg string) string {
	colorExp := regexp.MustCompile("\033\\[[0-9]+m")
	return colorExp.ReplaceAllString(msg, "")
}

// ExtractErrorMessage extract error from log
func ExtractErrorMessage(msg string) string {
	errExp := regexp.MustCompile(" error=\"([^\"]+)\"")
	if strings.Contains(msg, " ERR ") && errExp.MatchString(msg) {
		return errExp.FindStringSubmatch(msg)[1]
	}
	return ""
}
