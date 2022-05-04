package util

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	fs "github.com/fsnotify/fsnotify"
	ps "github.com/mitchellh/go-ps"
)

// TimeDifference seconds between remote and local time
var TimeDifference int64 = 0

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

// IsProcessExist check whether specified process still running
func IsProcessExist(pid int) bool {
	proc, err := ps.FindProcess(pid)
	if proc == nil || err != nil {
		return false
	}
	return strings.Contains(proc.Executable(), "ktctl")
}

// CreateDirIfNotExist create dir
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// WritePidFile write pid to file
func WritePidFile(componentName string, ch chan os.Signal) error {
	pidFile := fmt.Sprintf("%s/%s-%d.pid", KtHome, componentName, os.Getpid())
	if err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return err
	}
	go watchPidFile(pidFile, ch)
	return nil
}

func watchPidFile(pidFile string, ch chan os.Signal) {
	watcher, err := fs.NewWatcher()
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to create pid file watcher")
	}
	defer watcher.Close()

	if err = watcher.Add(pidFile); err != nil {
		log.Warn().Err(err).Msgf("Unable to watch pid file")
	}

	for event := range watcher.Events {
		log.Debug().Msgf("Received event %s", event)
		if event.Op & fs.Remove == fs.Remove || event.Op & fs.Rename == fs.Rename {
			log.Info().Msgf("Pid file was removed")
			ch <-os.Interrupt
		}
	}
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

// GetTime get time with rectification
func GetTime() int64 {
	return time.Now().Unix() + TimeDifference
}

// GetTimestamp get current timestamp
func GetTimestamp() string {
	return strconv.FormatInt(GetTime(), 10)
}

// ParseTimestamp parse timestamp to unix time
func ParseTimestamp(timestamp string) int64 {
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return -1
	}
	return unixTime
}

// FormattedTime get timestamp to print
func FormattedTime() string {
	return time.Now().Format(common.YyyyMmDdHhMmSs)
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
