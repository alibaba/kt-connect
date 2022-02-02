package util

import "github.com/alibaba/kt-connect/pkg/common"

func init() {
	_ = CreateDirIfNotExist(common.KtHome)
	FixFileOwner(common.KtHome)
}
