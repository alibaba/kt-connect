package util

import (
	"fmt"
	"os"
)

const HostsFilePath = "${SystemRoot}/System32/drivers/etc/hosts"
const Eol = "\r\n"

var UserHome = os.Getenv("USERPROFILE")
var KtHome = fmt.Sprintf("%s/.ktctl", UserHome)
