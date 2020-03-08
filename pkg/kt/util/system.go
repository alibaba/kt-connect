package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/lextoumbourou/goodhosts"
	"github.com/rs/zerolog/log"
)

var (
	pk []byte
)

// PrepareSSHPrivateKey generator ssh private key
func PrepareSSHPrivateKey(generator *SSHGenerator) (err error) {
	if len(generator.PrivateKey) == 0 {
		err = errors.New("private key must be set")
		return
	}
	err = ioutil.WriteFile(PrivateKeyPath(), generator.PrivateKey, 400)
	if err != nil {
		log.Error().Msgf("Fails create temp ssh private key")
	}
	return
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
