package cmd

import (
	"fmt"

	"github.com/alibaba/kt-connect/pkg/kt/util"
)

var (
	userHome = util.HomeDir()
	appHome  = fmt.Sprintf("%s/.ktctl", userHome)
	pidFile  = fmt.Sprintf("%s/pid", appHome)
)

func init() {
	util.CreateDirIfNotExist(appHome)
}
