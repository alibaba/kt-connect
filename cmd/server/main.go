package main

import (
	"github.com/alibaba/kt-connect/pkg/apiserver/cluster"
	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/alibaba/kt-connect/pkg/apiserver/server"
	"github.com/alibaba/kt-connect/pkg/apiserver/util"
	v1 "k8s.io/client-go/listers/core/v1"
)

var (
	podListener v1.PodLister
)

func main() {

	client, config, err := util.GetKubernetesClient()
	if err != nil {
		panic(err.Error())
	}

	watcher, err := cluster.Construct(client, config)
	if err != nil {
		panic(err.Error())
	}

	context := common.Context{
		Cluster: watcher,
	}

	server.Init(context)
}
