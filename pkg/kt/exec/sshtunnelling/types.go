package sshtunnelling

import "os/exec"

// CliInterface ...
type CliInterface interface {
	AddRoute(cidr string) *exec.Cmd
	AddDevice() *exec.Cmd
	RemoveDevice() *exec.Cmd
	SetupDeviceIP() *exec.Cmd
}

// Cli ...
type Cli struct {
	TunName  string
	SourceIP string
	DestIP   string
}
