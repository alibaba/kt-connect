package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
func IsDaemonRunning(componentName string) bool {
	files, _ := ioutil.ReadDir(KtHome)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), componentName) && strings.HasSuffix(f.Name(), ".pid") {
			return true
		}
	}
	return false
}

// IsPidFileExist check pid file is exist or not
func IsPidFileExist() bool {
	files, _ := ioutil.ReadDir(KtHome)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), fmt.Sprintf("-%d.pid", os.Getpid())) {
			return true
		}
	}
	return false
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
func WritePidFile(componentName string) error {
	pidFile := fmt.Sprintf("%s/%s-%d.pid", KtHome, componentName, os.Getpid())
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}
