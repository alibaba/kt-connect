//go:build !windows

package util

import (
	"os"
)

const (
	ResolvConf = "/etc/resolv.conf"
	FieldNameserver = "nameserver"
	FieldDomain = "domain"
	FieldSearch = "search"

	HostsFilePath = "/etc/hosts"
	Eol = "\n"
)

var (
	UserHome = os.Getenv("HOME")
)
