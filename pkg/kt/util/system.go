package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/lextoumbourou/goodhosts"
	"github.com/rs/zerolog/log"
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

// HomeDir Current User home dir
func HomeDir() string {
	// linux & mac
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	// windows
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	return "/root"
}

// KubeConfig location of kube-config file
func KubeConfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		kubeconfig = filepath.Join(HomeDir(), ".kube", "config")
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

// DropHosts ...
func DropHosts(hostsMap map[string]string) {
	hosts, err := goodhosts.NewHosts()

	if err != nil {
		log.Warn().Msgf("Fail to read hosts from host %s, ignore", err.Error())
		return
	}

	for name, ip := range hostsMap {
		if hosts.Has(ip, name) {
			if err = hosts.Remove(ip, name); err != nil {
				log.Warn().Str("ip", ip).Str("name", name).Msg("remove host failed")
			}
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Error().Err(err).Msgf("Error Happen when flush hosts")
	}

	log.Info().Msgf("- drop hosts successful.")
}

// DumpHosts DumpToHosts
func DumpHosts(hostsMap map[string]string) {
	hosts, err := goodhosts.NewHosts()

	if err != nil {
		log.Warn().Msgf("Fail to read hosts from host %s, ignore", err.Error())
		return
	}

	for name, ip := range hostsMap {
		if !hosts.Has(ip, name) {
			if err = hosts.Add(ip, name); err != nil {
				log.Warn().Str("ip", ip).Str("name", name).Msg("add host failed")
			}
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Error().Err(err).Msg("Error Happen when dump hosts")
	}

	log.Info().Msg("Dump hosts successful.")

}
