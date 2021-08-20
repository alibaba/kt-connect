package portforward

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

// Request ...
type Request struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// PodName pod name
	PodName string
	// Namespace target namespace
	Namespace string
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
	// Timeout connect timeout
	Timeout int
}

// ForwardPodPortToLocal ...
func (s *Cli) ForwardPodPortToLocal(request Request) (chan struct{}, context.Context, error) {
	stop := make(chan struct{})
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		process.Stop(<-stop, cancel)
	}()

	request.StopCh = stop

	go func() {
		err := portForward(request)
		if err != nil {
			stop <- struct{}{}
		}
	}()

	if !util.WaitPortBeReady(request.Timeout, request.LocalPort) {
		return nil, nil, errors.New("connect to port-forward failed")
	}
	util.SetupPortForwardHeartBeat(request.LocalPort)
	return stop, rootCtx, nil
}

// PortForward ...
func portForward(req Request) error {
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", req.Namespace, req.PodName)
	log.Debug().Msgf("Request port forward to %s", req.RestConfig.Host)
	apiUrl, err := parseReqHost(req.RestConfig.Host, apiPath)

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}
	fw, err := portforward.New(dialer, ports, req.StopCh, req.ReadyCh, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// parseReqHost get the final url to port forward api
func parseReqHost(host, apiPath string) (*url.URL, error) {
	pos := strings.Index(host, "://")
	if pos < 0 {
		return nil, fmt.Errorf("invalid host address: %s", host)
	}
	protocol := host[0:pos]
	hostIP := host[pos+3:]
	baseUrl := ""
	pos = strings.Index(hostIP, "/")
	if pos > 0 {
		baseUrl = hostIP[pos:]
		hostIP = hostIP[0:pos]
	}
	fullPath := path.Join(baseUrl, apiPath)
	return &url.URL{Scheme: protocol, Host: hostIP, Path: fullPath}, nil
}
