// +build !windows

package util

import (
	"fmt"
	"os"
)

const hostsFilePath = "/etc/hosts"
const eol = "\n"

var UserHome = os.Getenv("HOME")
var KtHome = fmt.Sprintf("%s/.ktctl", UserHome)
