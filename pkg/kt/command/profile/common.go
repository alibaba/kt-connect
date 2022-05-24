package profile

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
)

func profileFile(profile string) string {
	return fmt.Sprintf("%s/%s", util.KtProfileDir, profile)
}
