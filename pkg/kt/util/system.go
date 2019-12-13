package util

import (
	"os"
	"fmt"
)

// IsDaemonRunning check daemon is running or not
func IsDaemonRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// HomeDir Current User home dir
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	return "/root"
}

// PrivateKeyPath Get ssh private key path
func PrivateKeyPath() string {
	userHome := HomeDir()
	privateKey := fmt.Sprintf("%s/.kt_id_rsa", userHome)
	return privateKey
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}