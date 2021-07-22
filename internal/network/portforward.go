package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alibaba/kt-connect/internal/process"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForwardAPodRequest ...
type PortForwardAPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// PodName pod name
	PodName string
	// Namespace namepsace
	Namespace string
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
	Timeout int
}

// ForwardPodPortToLocal ...
func ForwardPodPortToLocal(request PortForwardAPodRequest) (chan struct{}, context.Context, error) {
	stop := make(chan struct{})
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		process.Stop(<-stop, cancel)
	}()

	request.StopCh = stop

	go func() {
		err := PortForward(request)
		if err != nil {
			stop <- struct{}{}
		}
	}()

	ready := waitPortBeReady(request.Timeout, request.LocalPort)

	if !ready {
		return nil, nil, errors.New("port-forward not ready")
	}
	return stop, rootCtx, nil
}

// PortForward ...
func PortForward(req PortForwardAPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", req.Namespace, req.PodName)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// waitPortBeReady return true when port is ready
// It waits at most waitTime seconds, then return false.
func waitPortBeReady(waitTime, port int) bool {
	for i := 0; i < waitTime; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Debug().Msgf("connect to port-forward failed, error: %s, retry: %d", err, i)
			time.Sleep(1 * time.Second)
		} else {
			conn.Close()
			log.Info().Msgf("connect to port-forward successful")
			return true
		}
	}
	return false
}
