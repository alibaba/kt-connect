package controllers

import (
	context2 "context"
	"fmt"
	"net/http"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	networking "istio.io/api/networking/v1alpha3"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IstioController istio api controller
type IstioController struct {
	Context common.Context
}

// VirtualServices list virtual service
func (c *IstioController) VirtualServices(context *gin.Context) {
	namespace := context.Param("namespace")

	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create istio client")
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	vsList, err := ic.NetworkingV1alpha3().VirtualServices(namespace).List(context2.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get VirtualService in %s namespace", namespace)
	}

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	context.JSON(http.StatusOK, vsList)
}

// VirtualService get virtual service instance
func (c *IstioController) VirtualService(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")

	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	vs, err := ic.NetworkingV1alpha3().VirtualServices(namespace).Get(context2.TODO(), name, metav1.GetOptions{})

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get virtual service",
		})
		return
	}

	context.JSON(http.StatusOK, vs)
}

// DestinationRules get destination rule
func (c *IstioController) DestinationRules(context *gin.Context) {
	namespace := context.Param("namespace")
	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	destinationrules, err := ic.NetworkingV1alpha3().DestinationRules(namespace).List(context2.TODO(), metav1.ListOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get destinationrules",
		})
		return
	}

	context.JSON(http.StatusOK, destinationrules)
}

// DestinationRule get destination rule instances
func (c *IstioController) DestinationRule(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	destinationrule, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Get(context2.TODO(), name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("fail get destinationrule %s", name),
		})
		return
	}

	context.JSON(http.StatusOK, destinationrule)
}

// AddVersionToDestinationRule add version to destination rule
func (c *IstioController) AddVersionToDestinationRule(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	version := context.Param("version")

	ic, err := versionedclient.NewForConfig(c.Context.Cluster.Config)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail init istioClient",
		})
		return
	}

	destinationrule, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Get(context2.TODO(), name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get destinationrule",
		})
		return
	}

	for _, subset := range destinationrule.Spec.Subsets {
		if subset.Name == version {
			context.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "version already present",
			})
			return
		}
	}

	newSubset := &networking.Subset{
		Name: version,
		Labels: map[string]string{
			version: version,
		},
	}
	destinationrule.Spec.Subsets = append(destinationrule.Spec.Subsets, newSubset)

	result, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Update(context2.TODO(), destinationrule, metav1.UpdateOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, result)
}

// RemoveVersionToDestinationRule remove version from destination rule
func (c *IstioController) RemoveVersionToDestinationRule(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "remove version to destinaltion rule",
	})
}
