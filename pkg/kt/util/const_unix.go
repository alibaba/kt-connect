//go:build !windows

package util

import (
	"fmt"
	"os"
)

const ResolvConf = "/etc/resolv.conf"
const FieldNameserver = "nameserver"
const FieldDomain = "domain"
const FieldSearch = "search"

const HostsFilePath = "/etc/hosts"
const Eol = "\n"

var UserHome = os.Getenv("HOME")
var KtHome = fmt.Sprintf("%s/.ktctl", UserHome)
