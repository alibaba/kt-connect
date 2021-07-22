package controllers

import (
	"net/http"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	ktConst "github.com/alibaba/kt-connect/pkg/common"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/labels"
)

// KTController KTController
type KTController struct {
	Context common.Context
}

// Components Components
func (c KTController) Components(context *gin.Context) {
	set := labels.Set{
		ktConst.ControlBy: ktConst.KubernetesTool,
	}
	selector := labels.SelectorFromSet(set)
	pods, err := c.Context.Cluster.PodLister.List(selector)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(http.StatusOK, pods)
}

// ComponentsInNamespace ComponentsInNamespace
func (c KTController) ComponentsInNamespace(context *gin.Context) {
	namespace := context.Param("namespace")
	set := labels.Set{
		ktConst.ControlBy: ktConst.KubernetesTool,
	}
	selector := labels.SelectorFromSet(set)
	pods, err := c.Context.Cluster.PodLister.Pods(namespace).List(selector)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(http.StatusOK, pods)
}
