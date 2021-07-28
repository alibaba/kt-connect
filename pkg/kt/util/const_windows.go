package util

import (
	"fmt"
	"os"
)

const hostsFilePath = "${SystemRoot}/System32/drivers/etc/hosts"
const eol = "\r\n"

var UserHome = os.Getenv("USERPROFILE")
var KtHome = fmt.Sprintf("%s/.ktctl", UserHome)
