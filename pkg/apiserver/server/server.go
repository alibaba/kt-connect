package server

import "github.com/alibaba/kt-connect/pkg/apiserver/common"

func Init(context common.Context) {
	r := NewRouter(context)
	r.Run(":8000")
}
