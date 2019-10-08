package controllers

import (
	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type IstioController struct {
	Context common.Context
}

func (c *IstioController) VirtualServices(context *gin.Context) {
	namespace := context.Param("namespace")
	dynamicClient, err := dynamic.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init dynamicClient",
		})
		return
	}

	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}

	virtualServices, err := dynamicClient.Resource(virtualServiceGVR).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get pod log",
		})
		return
	}

	context.JSON(200, virtualServices)
}

func (c *IstioController) VirtualService(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	dynamicClient, err := dynamic.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init dynamicClient",
		})
		return
	}

	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}

	virtualService, err := dynamicClient.Resource(virtualServiceGVR).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get pod log",
		})
		return
	}

	context.JSON(200, virtualService)
}

func (c *IstioController) DestinationRules(context *gin.Context) {
	namespace := context.Param("namespace")
	dynamicClient, err := dynamic.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init dynamicClient",
		})
		return
	}

	destinationruleGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "destinationrules",
	}

	destinationrules, err := dynamicClient.Resource(destinationruleGVR).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get pod log",
		})
		return
	}

	context.JSON(200, destinationrules)
}

func (c *IstioController) DestinationRule(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	dynamicClient, _ := dynamic.NewForConfig(c.Context.Cluster.Config)

	destinationruleGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "destinationrules",
	}

	destinationrule, err := dynamicClient.Resource(destinationruleGVR).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get pod log",
		})
		return
	}

	context.JSON(200, destinationrule)
}
