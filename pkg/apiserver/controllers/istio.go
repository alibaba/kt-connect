package controllers

import (
	"log"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IstioController struct {
	Context common.Context
}

func (c *IstioController) VirtualServices(context *gin.Context) {
	namespace := context.Param("namespace")

	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}

	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	context.JSON(200, vsList)
}

func (c *IstioController) VirtualService(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")

	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	vs, err := ic.NetworkingV1alpha3().VirtualServices(namespace).Get(name, metav1.GetOptions{})

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get virtual service",
		})
		return
	}

	context.JSON(200, vs)
}

func (c *IstioController) DestinationRules(context *gin.Context) {
	namespace := context.Param("namespace")
	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	destinationrules, err := ic.NetworkingV1alpha3().DestinationRules(namespace).List(metav1.ListOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get destinationrules",
		})
		return
	}

	context.JSON(200, destinationrules)
}

func (c *IstioController) DestinationRule(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	destinationrule, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get destinationrule",
		})
		return
	}

	context.JSON(200, destinationrule)
}
