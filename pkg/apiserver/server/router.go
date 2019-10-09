package server

import (
	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/alibaba/kt-connect/pkg/apiserver/controllers"
	"github.com/gin-gonic/gin"
)

// NewRouter api router
func NewRouter(context common.Context) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	ktController := controllers.KTController{Context: context}
	clusterController := controllers.ClusterController{Context: context}
	terminalController := controllers.TerminalController{Context: context}
	istioController := controllers.IstioController{Context: context}

	router.GET("/api/components", ktController.Components)
	router.GET("/api/cluster/namespaces", clusterController.Namespaces)
	router.GET("/api/cluster/namespaces/:namespace/components", ktController.ComponentsInNamespace)
	router.GET("/api/cluster/namespaces/:namespace/services", clusterController.Services)
	router.GET("/api/cluster/namespaces/:namespace/services/:name", clusterController.Service)
	router.GET("/api/cluster/namespaces/:namespace/endpoints", clusterController.Endpoints)
	router.GET("/api/cluster/namespaces/:namespace/endpoints/:name", clusterController.Endpoint)
	router.GET("/api/cluster/namespaces/:namespace/replicasets/:name", clusterController.ReplicaSet)
	router.GET("/api/cluster/namespaces/:namespace/deployments", clusterController.Deployments)
	router.GET("/api/cluster/namespaces/:namespace/deployments/:name", clusterController.Deployment)
	router.GET("/api/cluster/namespaces/:namespace/pods", clusterController.Pods)
	router.GET("/api/cluster/namespaces/:namespace/pods/:name", clusterController.Pod)
	router.GET("/api/cluster/namespaces/:namespace/pods/:name/log", clusterController.PodLog)

	router.GET("/api/cluster/namespaces/:namespace/virtualservices", istioController.VirtualServices)
	router.GET("/api/cluster/namespaces/:namespace/virtualservices/:name", istioController.VirtualService)
	router.GET("/api/cluster/namespaces/:namespace/destinationrules", istioController.DestinationRules)
	router.GET("/api/cluster/namespaces/:namespace/destinationrules/:name", istioController.DestinationRule)

	router.GET("/ws/terminal", terminalController.Terminal)

	return router
}
