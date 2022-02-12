//go:build !windows

package util

import (
	"os"
)

func IsRunAsAdmin() bool {
	return os.Geteuid() == 0
}

func GetAdminUserName() string {
	return "root"
}
