package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
)

// GetDaemonRunning fetch daemon pid if exist
func GetDaemonRunning(componentName string) int {
	files, _ := ioutil.ReadDir(KtHome)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), componentName) && strings.HasSuffix(f.Name(), ".pid") {
			from := len(componentName) + 1
			to := len(f.Name()) - len(".pid")
			pid, err := strconv.Atoi(f.Name()[from:to])
			if err == nil && IsProcessExist(pid) {
				return pid
			}
		}
	}
	return -1
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

// FixFileOwner set owner to original user when run with sudo
func FixFileOwner(path string) {
	var uid int
	var gid int
	sudoUid := os.Getenv("SUDO_UID")
	if sudoUid == "" {
		uid = os.Getuid()
	} else {
		uid, _ = strconv.Atoi(sudoUid)
	}
	sudoGid := os.Getenv("SUDO_GID")
	if sudoGid == "" {
		gid = os.Getuid()
	} else {
		gid, _ = strconv.Atoi(sudoGid)
	}
	_ = os.Chown(path, uid, gid)
}

// GetTimestamp get current time stamp
func GetTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// GetLocalUserName get current username
func GetLocalUserName() string {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		return sudoUser
	}
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

// IsCmd check running in windows cmd shell
func IsCmd() bool {
	proc, _ := ps.FindProcess(os.Getppid())
	if proc != nil && !strings.Contains(proc.Executable(), "cmd.exe") {
		return false
	}
	return true
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux check runtime is windows
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
