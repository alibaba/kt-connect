package controllers

import (
	"fmt"
	"net/http"

	"github.com/alibaba/kt-connect/pkg/apiserver/common"
	"github.com/alibaba/kt-connect/pkg/apiserver/ws"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// TerminalController terminal rest api
type TerminalController struct {
	Context common.Context
}

// Terminal ...
func (c *TerminalController) Terminal(context *gin.Context) {
	wsConn, err := ws.Constructor(context.Writer, context.Request)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail create websocket",
		})
		return
	}

	ns := context.Query("ns")
	pod := context.Query("p")
	container := context.Query("c")
	shell := context.Query("s")
	if shell == "" {
		shell = "bash"
	}

	client := c.Context.Client()
	config := c.Context.Config()

	sshReq := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{shell},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		},
			scheme.ParameterCodec,
		)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", sshReq.URL())

	if err != nil {
		fmt.Printf("error happen when connect pod: %s", err.Error())
		return
	}

	handler := &ws.StreamHandler{WsConn: wsConn, ResizeEvent: make(chan remotecommand.TerminalSize)}

	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	})

	if err != nil {
		fmt.Printf("error happen when executor get stream: %s", err.Error())
		return
	}

}
