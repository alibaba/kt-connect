package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/gin-gonic/gin"
	"github.com/kubernetes/dashboard/src/app/backend/resource/container"
	"github.com/kubernetes/dashboard/src/app/backend/resource/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ClusterController struct {
	Context common.Context
}

func (c *ClusterController) Namespaces(context *gin.Context) {
	namespaces, err := c.Context.NamespaceLister().List(labels.Everything())
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list namespace",
		})
		return
	}
	context.JSON(200, namespaces)
}

func (c *ClusterController) Services(context *gin.Context) {
	namespace := context.Param("namespace")
	services, err := c.Context.ServiceLister().Services(namespace).List(labels.Everything())
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list service",
		})
		return
	}
	context.JSON(200, services)
}

func (c *ClusterController) Service(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	service, err := c.Context.Client().CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get service",
		})
		return
	}
	context.JSON(200, service)
}

func (c *ClusterController) Endpoints(context *gin.Context) {
	namespace := context.Param("namespace")
	endpoints, err := c.Context.EndpointsLister().Endpoints(namespace).List(labels.Everything())
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list service",
		})
		return
	}
	context.JSON(200, endpoints)
}

func (c *ClusterController) Endpoint(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	endpoint, err := c.Context.Client().CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get endpoint",
		})
		return
	}
	context.JSON(200, endpoint)
}

func (c *ClusterController) Deployments(context *gin.Context) {
	namespace := context.Param("namespace")
	selector := context.Query("selector")

	options := metav1.ListOptions{}

	if selector != "" {
		var m labels.Set
		err := json.Unmarshal([]byte(selector), &m)
		if err == nil {
			var labelSelector labels.Selector
			labelSelector = labels.SelectorFromSet(m)
			options.FieldSelector = fmt.Sprintf("spec.selector.matchLabels=%s", labelSelector.String())
		}

	}

	resource, err := c.Context.Client().ExtensionsV1beta1().Deployments(namespace).List(options)

	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get deployment",
		})
		return
	}
	context.JSON(200, resource)
}

func (c *ClusterController) Deployment(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	resource, err := c.Context.Client().ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get deployment",
		})
		return
	}
	context.JSON(200, resource)
}

func (c *ClusterController) ReplicaSet(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")
	resource, err := c.Context.Client().ExtensionsV1beta1().ReplicaSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get replicaSet",
		})
		return
	}
	context.JSON(200, resource)
}

func (c *ClusterController) Pods(context *gin.Context) {
	namespace := context.Param("namespace")
	selector := context.Query("selector")
	var m labels.Set
	err := json.Unmarshal([]byte(selector), &m)
	var labelSelector labels.Selector
	if m != nil {
		fmt.Println("query by set")
		labelSelector = labels.SelectorFromSet(m)
	} else {
		fmt.Println("everything")
		labelSelector = labels.Everything()
	}

	pods, err := c.Context.PodLister().Pods(namespace).List(labelSelector)
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail list pods",
		})
		return
	}
	context.JSON(200, pods)
}

func (c *ClusterController) Pod(context *gin.Context) {
	namespace := context.Param("namespace")
	name := context.Param("name")

	pod, err := c.Context.Client().CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		context.JSON(500, gin.H{
			"message": "fail get pod",
		})
		return
	}
	context.JSON(200, pod)
}

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
		context.JSON(500, gin.H{
			"message": "fail get pod log",
		})
		return
	}

	context.JSON(200, result)
}
