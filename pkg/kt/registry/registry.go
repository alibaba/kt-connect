// +build !windows

package registry

import "fmt"

func SetGlobalProxy(port int) error {
	if port > 0 {
		return nil
	}
	return fmt.Errorf("invalid socks port %d", port)
}

func CleanGlobalProxy() {
}
