package transmission

import (
	"context"
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/process"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
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

// SetupPortForwardToLocal mapping local port to shadow pod ssh port
func SetupPortForwardToLocal(podName string, remotePort, localPort int) (chan struct{}, error) {
	stop := make(chan struct{})
	_, cancel := context.WithCancel(context.Background())
	// one of the background process start failed and will cancel the started process
	go func() {
		process.Stop(<-stop, cancel)
	}()
	go func() {
		if err := portForward(podName, remotePort, localPort, stop); err != nil {
			log.Error().Err(err).Msgf("Port forward to %d -> %d pod %s interrupted", localPort, remotePort, podName)
			stop <- struct{}{}
		}
	}()

	if !util.WaitPortBeReady(opt.Get().PortForwardWaitTime, localPort) {
		return nil, fmt.Errorf("connect to port-forward failed")
	}
	cluster.SetupPortForwardHeartBeat(localPort)
	return stop, nil
}

// PortForward call port forward api
func portForward(podName string, remotePort, localPort int, stop chan struct{}) error {
	ready := make(chan struct{})
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", opt.Get().Namespace, podName)
	log.Debug().Msgf("Request port forward pod:%d -> local:%d via %s", remotePort, localPort, opt.Get().RuntimeStore.RestConfig.Host)
	apiUrl, err := parseReqHost(opt.Get().RuntimeStore.RestConfig.Host, apiPath)

	transport, upgrader, err := spdy.RoundTripperFor(opt.Get().RuntimeStore.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	var out io.Writer = nil
	if opt.Get().Debug {
		out = os.Stdout
	}
	fw, err := portforward.New(dialer, ports, stop, ready, out, os.Stderr)
	if err != nil {
		return err
	}
	// in client-go 0.22, this always return nil
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
