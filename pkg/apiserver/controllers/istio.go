package controllers

import (
	"log"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	v1alpha3 "istio.io/api/networking/v1alpha3"
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

func (c *IstioController) AddVersionToDestinationRule(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	version := context.Param("version")

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

	for _, subset := range destinationrule.Spec.Subsets {
		if subset.Name == version {
			context.JSON(422, gin.H{
				"message": "version already present",
			})
			return
		}
	}

	newSubset := &v1alpha3.Subset{
		Name: version,
		Labels: map[string]string{
			version: version,
		},
	}
	destinationrule.Spec.Subsets = append(destinationrule.Spec.Subsets, newSubset)

	result, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Update(destinationrule)
	if err != nil {
		context.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	context.JSON(200, result)
}

func (c *IstioController) RemoveVersionToDestinationRule(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "remove version to destinaltion rule",
	})
}
