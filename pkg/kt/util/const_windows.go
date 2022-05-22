package util

import (
	"os"
)

const (
	HostsFilePath = "${SystemRoot}/System32/drivers/etc/hosts"
	Eol = "\r\n"
)

var (
	UserHome = os.Getenv("USERPROFILE")
)

