package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

// GetDaemonRunning fetch daemon pid if exist
func GetDaemonRunning(componentName string) string {
	files, _ := ioutil.ReadDir(KtHome)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), componentName) && strings.HasSuffix(f.Name(), ".pid") {
			from := len(componentName) + 1
			to := len(f.Name()) - len(".pid")
			return f.Name()[from:to]
		}
	}
	return ""
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

// IsProcessExist check whether specified process still running
func IsProcessExist(pid int) bool {
	proc, err := ps.FindProcess(pid)
	if proc == nil || err != nil {
		return false
	}
	return strings.Contains(proc.Executable(), "ktctl")
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

// GetJvmrcFilePath get jvmrc file from jvmrc dir
func GetJvmrcFilePath(jvmrcDir string) string {
	if jvmrcDir != "" {
		folder, err := os.Stat(jvmrcDir)
		if err == nil && folder.IsDir() {
			return filepath.Join(jvmrcDir, ".jvmrc")
		}
	}
	return ""
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux check runtime is windows
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
