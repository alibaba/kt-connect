package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var interrupt = make(chan bool)

// StopBackendProcess ...
func StopBackendProcess(stop bool, cancel func()) {
	if cancel == nil {
		return
	}
	cancel()
	interrupt <- stop
}

// Interrupt ...
func Interrupt() chan bool {
	return interrupt
}

// IsDaemonRunning check daemon is running or not
func IsDaemonRunning(pidFile string) bool {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// KubeConfig location of kube-config file
func KubeConfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		kubeconfig = filepath.Join(UserHome, ".kube", "config")
	}
	return kubeconfig
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

// WritePidFile write pid to file
func WritePidFile(pidFile string) (pid int, err error) {
	pid = os.Getpid()
	err = ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	return
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}
