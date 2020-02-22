package integration

import (
	"time"
)

// Minutes will return timeout in minutes based on how slow the machine is
func Minutes(n int) time.Duration {
	return time.Duration(*timeOutMultiplier) * time.Duration(n) * time.Minute
}
