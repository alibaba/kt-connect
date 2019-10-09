package controllers

import (
	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/labels"
)

type KTController struct {
	Context common.Context
}

func (c KTController) Components(context *gin.Context) {
	set := labels.Set{
		"control-by": "kt",
	}
	selector := labels.SelectorFromSet(set)
	pods, err := c.Context.Cluster.PodLister.List(selector)
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(200, pods)
}

func (c KTController) ComponentsInNamespace(context *gin.Context) {
	namespace := context.Param("namespace")
	set := labels.Set{
		"control-by": "kt",
	}
	selector := labels.SelectorFromSet(set)
	pods, err := c.Context.Cluster.PodLister.Pods(namespace).List(selector)
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(200, pods)
}
