package portforward

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/alibaba/kt-connect/pkg/process"
	"github.com/rs/zerolog/log"
	"io"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

// ForwardPodPortToLocal ...
func (s *Cli) ForwardPodPortToLocal(options *options.DaemonOptions, podName string, remotePort, localPort int) (chan struct{}, context.Context, error) {

	// If localSSHPort is in use by another process, return an error.
	ready := util.WaitPortBeReady(1, localPort)
	if ready {
		return nil, nil, fmt.Errorf("127.0.0.1:%d already in use", localPort)
	}

	stop := make(chan struct{})
	rootCtx, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		process.Stop(<-stop, cancel)
	}()
	go func() {
		err := portForward(options, podName, remotePort, localPort, stop)
		if err != nil {
			stop <- struct{}{}
		}
	}()

	if !util.WaitPortBeReady(options.WaitTime, localPort) {
		return nil, nil, errors.New("connect to port-forward failed")
	}
	util.SetupPortForwardHeartBeat(localPort)
	return stop, rootCtx, nil
}

// PortForward ...
func portForward(options *options.DaemonOptions, podName string, remotePort, localPort int, stop chan struct{}) error {
	ready := make(chan struct{})
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", options.Namespace, podName)
	log.Debug().Msgf("Request port forward to %s", options.RuntimeOptions.RestConfig.Host)
	apiUrl, err := parseReqHost(options.RuntimeOptions.RestConfig.Host, apiPath)

	transport, upgrader, err := spdy.RoundTripperFor(options.RuntimeOptions.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	var out io.Writer = nil
	if options.Debug {
		out = os.Stdout
	}
	fw, err := portforward.New(dialer, ports, stop, ready, out, os.Stderr)
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
