package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	"github.com/kubernetes/dashboard/src/app/backend/resource/container"
	"github.com/kubernetes/dashboard/src/app/backend/resource/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ClusterController provide kubernetes cluster api
type ClusterController struct {
	Context common.Context
}

// Namespaces list namespaces
func (c *ClusterController) Namespaces(context *gin.Context) {
	namespaces, err := c.Context.NamespaceLister().List(labels.Everything())
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list namespace",
		})
		return
	}
	context.JSON(http.StatusOK, namespaces)
}

// Services list services
func (c *ClusterController) Services(context *gin.Context) {
	namespace := context.Param("namespace")
	services, err := c.Context.ServiceLister().Services(namespace).List(labels.Everything())
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list service",
		})
		return
	}
	context.JSON(http.StatusOK, services)
}

// Service get service instance
func (c *ClusterController) Service(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	service, err := c.Context.Client().CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get service " + name,
		})
		return
	}
	context.JSON(http.StatusOK, service)
}

// Endpoints list endpoints
func (c *ClusterController) Endpoints(context *gin.Context) {
	namespace := context.Param("namespace")
	endpoints, err := c.Context.EndpointsLister().Endpoints(namespace).List(labels.Everything())
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list service",
		})
		return
	}
	context.JSON(http.StatusOK, endpoints)
}

// Endpoint get endpoint instance
func (c *ClusterController) Endpoint(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	endpoint, err := c.Context.Client().CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get endpoint",
		})
		return
	}
	context.JSON(http.StatusOK, endpoint)
}

// Deployments list deployments
func (c *ClusterController) Deployments(context *gin.Context) {
	namespace := context.Param("namespace")
	selector := context.Query("selector")
	options := metav1.ListOptions{}

	if selector != "" {
		labelSelector, err := querySelector(selector)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"message": "bad request",
			})
			return
		}
		options.FieldSelector = fmt.Sprintf("spec.selector.matchLabels=%s", labelSelector.String())
	}

	resource, err := c.Context.Client().AppsV1().Deployments(namespace).List(options)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get deployment",
		})
		return
	}
	context.JSON(http.StatusOK, resource)
}

// Deployment get deployment instance
func (c *ClusterController) Deployment(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	resource, err := c.Context.Client().AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get deployment",
		})
		return
	}
	context.JSON(http.StatusOK, resource)
}

// ReplicaSet get replicaSet instance
func (c *ClusterController) ReplicaSet(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	resource, err := c.Context.Client().AppsV1().ReplicaSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get replicaSet",
		})
		return
	}
	context.JSON(http.StatusOK, resource)
}

// Pods list pods
func (c *ClusterController) Pods(context *gin.Context) {
	namespace := context.Param("namespace")
	selector := context.Query("selector")
	labelSelector, err := querySelector(selector, labels.Everything())
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
	}

	pods, err := c.Context.PodLister().Pods(namespace).List(labelSelector)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(http.StatusOK, pods)
}

// Pod get pod instance
func (c *ClusterController) Pod(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")

	pod, err := c.Context.Client().CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail get pod",
		})
		return
	}
	context.JSON(http.StatusOK, pod)
}

// PodLog get pod log
func (c *ClusterController) PodLog(context *gin.Context) {
	namespace := context.Param("namespace")
	podID := context.Param("name")
	containerID := context.Query("container")
	refTimestamp := context.Query("referenceTimestamp")

	if refTimestamp == "" {
		refTimestamp = logs.NewestTimestamp
	}

	refLineNum, err := strconv.Atoi(context.Query("referenceLineNum"))
	if err != nil {
		refLineNum = 0
	}

	usePreviousLogs := context.Query("previous") == "true"
	offsetFrom, err1 := strconv.Atoi(context.Query("offsetFrom"))
	offsetTo, err2 := strconv.Atoi(context.Query("offsetTo"))
	logFilePosition := context.Query("logFilePosition")

	logSelector := logs.DefaultSelection
	if err1 == nil && err2 == nil {
		logSelector = &logs.Selection{
			ReferencePoint: logs.LogLineId{
				LogTimestamp: logs.LogTimestamp(refTimestamp),
				LineNum:      refLineNum,
			},
			OffsetFrom:      offsetFrom,
			OffsetTo:        offsetTo,
			LogFilePosition: logFilePosition,
		}
	}

	result, err := container.GetLogDetails(c.Context.Client(), namespace, podID, containerID, logSelector, usePreviousLogs)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("fail get pod %s log", podID),
		})
		return
	}

	context.JSON(http.StatusOK, result)
}

func querySelector(selector string, defSelector ...labels.Selector) (labels.Selector, error) {
	set := make(labels.Set)
	if selector == "" && len(defSelector) != 0 {
		return defSelector[0], nil
	}
	if selector == "" {
		return set.AsSelector(), nil
	}
	if err := json.Unmarshal([]byte(selector), &set); err != nil {
		return nil, err
	}
	return labels.SelectorFromSet(set), nil
}
