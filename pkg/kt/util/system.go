package util

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"

	"time"

	"github.com/rs/zerolog/log"
	"github.com/lextoumbourou/goodhosts"
)

var (
	pk []byte
)

func init() {
	rand.Seed(time.Now().UnixNano())
	pk = []byte("-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIIEpAIBAAKCAQEAvSVAezJDBhrNDhLhuCaKrvdtCFTdqJmGLGyfBqEYb3p4a91g\n" +
		"l4gD2LGwiRlgpwU4oSECeMmwP53C4vrfPKY45/+8lncIkE/4E+8hHssXkHaqQrjE\n" +
		"FtePfxZ6/xi84kUbWNNV4IGAeFwXtq9GszQ+kWMNT5QmuexOXOlqq7W4CIAUe3uX\n" +
		"29WCp3OGiBeP4ORDraRa/1bwBH+Cq0UxEYT+6EuDU0YzF3JF4H8At6NdgElAuezE\n" +
		"wI84p5LNr1HmTPndcHJtX2+POKEoNYBxPekEyJbqExIR2dLRUytlX5tIacKkMBCJ\n" +
		"aX7DBtJzWX7BxfjRfXjzNOpufuTU9BsknyJlFwIDAQABAoIBAQCS606s4xvAsCy7\n" +
		"U9tUyUtMIRDmOdV7UtUvyKe15Igwf3bugiS3T4V9Wnh/5eB3m8yjDBr5a+ClaYup\n" +
		"96hTWeI2AyWf0pIqVpOiGEsnuiVxp1sVPKPEAmiKFRIw+CwvrfJSCsZX/v+lfhNF\n" +
		"adyG8nvvPntmZvO102IDNaQQALUUk+J69yvtb7ekfvZCSmanXx/R7y2+u8Hd8wtf\n" +
		"fpyVcM1g7YZwhxto2uKyUE1T/myOT5+wULfYTMNynsLB6dXJJP+a89I6cRdFCIib\n" +
		"kqoZ5FuaTXXucrgWGDne6DcvNbiZi1f8LRb6RFnlvv6D42xyjelyoGY7BktKsFwJ\n" +
		"NwLR1lBhAoGBANyGB7ti190DZoDGQvIpSHd+JeZHoK3g49VNxGFU+SAhXa8gQn6K\n" +
		"Xi5qNRD2XLTEnT36U20/bkcDv0oSTZikJhU0OqxgouVO3YZ2cZhXcoPZmf+vhgzI\n" +
		"ufv0T8/HlQyr7Sp9VqfqlC7u1P83VbNro/D3V+wtNKog4g++DvKtsh8JAoGBANuS\n" +
		"9XDaAnq5K0rjShNqCMRyHx6+kFPaLpL4kt1f4yra8w/m3pes63sO1vz/4wOhoalL\n" +
		"imAEqTKTblinPhjCxbe4e/WqnAQM05XROdiGer2RhBIMCo2/YE3WLCWAyVCDtd3B\n" +
		"Te9rPynSsAmtgDRuftusY7TAIuwZuG4K71Gw8UsfAoGAYmRm5MPYXqNaw9AyJIwo\n" +
		"6i/dxx5kYdB6tzxoh6j7MsvQWggBwyYHmZwHq1bQzFMBeZrMSG1JzeOtIOaDurxa\n" +
		"xZE1MJ45cCi9DHaifn9d99hKLtvo6qFQ4ksCpUl+hlXbjt63oFo43avwWyMcWN6J\n" +
		"GkWx9A3DdrkPREjfsIWxeMkCgYEAtDhv6duWk2IujX32y+6JGaxNrK9eyORYu9r4\n" +
		"uGi+jOs++ztUUgvlD5EDlo70poNgrBLLlbndohxuQqeqiSo8nGn4nJAXFB/u/pXH\n" +
		"M9hVIAky7JkjhGqiweBbRcDp+4LPoB7MOAm/wzUhth/JDb/vsaBSCgZ143HM9c1V\n" +
		"1qgztKMCgYAUhQRJB6ofGqiGsPN2KZw+0IoPNS3Tk0NTjzVh2o927B8zb0T0bO5e\n" +
		"qe0OO7FFGcON6uSOkGu2p9KHUEm6OFaQLjdysjrGI7GVRYW7D/SSLidRREv2A70R\n" +
		"f0/Mi8v9nD4ztroXQDeeL8O4rFTnfRdqs+MZ/MYoq9C5iE1IHJm7KQ==\n" +
		"-----END RSA PRIVATE KEY-----")
}

// PrepareSSHPrivateKey generator ssh private key
func PrepareSSHPrivateKey() (err error) {
	err = ioutil.WriteFile(PrivateKeyPath(), pk, 400)
	if err != nil {
		log.Printf("Fails create temp ssh private key")
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

// IsHelpCommand IsHelpCommand
func IsHelpCommand(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// IsWindows check runtime is windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DumpToHosts DumpToHosts
func DumpToHosts(hostsMap map[string]string) {
	hosts, err := goodhosts.NewHosts()

	if err !=nil {
		log.Printf("Fail to read hosts from host %s, ignore", err.Error())
		return
	}

	for name, ip := range hostsMap {
		if !hosts.Has(ip, name) {
			hosts.Add(ip, name)
		}
	}

	if err := hosts.Flush(); err != nil {
		log.Info().Msgf("Error Happen when flush hosts")
	}

}
