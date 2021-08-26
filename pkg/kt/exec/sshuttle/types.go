package sshuttle

import "os/exec"

// CliInterface ...
type CliInterface interface {
	Version() *exec.Cmd
	Install() *exec.Cmd
	Connect(remoteHost, privateKeyPath string, remotePort int, DNSServer string, disableDNS bool, cidrs []string, debug bool) *exec.Cmd
}

// Cli ...
type Cli struct{}
