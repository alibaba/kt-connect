package transmission

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/rs/zerolog/log"
	"io"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// SetupPortForwardToLocal mapping local port to shadow pod ssh port
func SetupPortForwardToLocal(podName string, remotePort, localPort int) error {
	ready := make(chan struct{})
	var ticker *time.Ticker
	go func() {
		if err := portForward(podName, remotePort, localPort, ready); err != nil {
			log.Error().Err(err).Msgf("Port forward to %d -> %d pod %s interrupted", localPort, remotePort, podName)
			time.Sleep(time.Duration(opt.Get().PortForwardWaitTime) * time.Second)
		} else {
			if ticker != nil {
				ticker.Stop()
			}
		}
		log.Debug().Msgf("Port forward reconnecting ...")
		_ = SetupPortForwardToLocal(podName, remotePort, localPort)
	}()

	select {
	case <-ready:
		ticker = cluster.SetupPortForwardHeartBeat(localPort)
	case <-time.After(time.Duration(opt.Get().PortForwardWaitTime) * time.Second):
		return fmt.Errorf("connect to port-forward failed")
	}
	return nil
}

// PortForward call port forward api
func portForward(podName string, remotePort, localPort int, ready chan struct{}) error {
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", opt.Get().Namespace, podName)
	log.Debug().Msgf("Request port forward pod:%d -> local:%d via %s", remotePort, localPort, opt.Get().RuntimeStore.RestConfig.Host)
	apiUrl, err := parseReqHost(opt.Get().RuntimeStore.RestConfig.Host, apiPath)

	transport, upgrader, err := spdy.RoundTripperFor(opt.Get().RuntimeStore.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	fw, err := portforward.New(dialer, ports, make(<-chan struct{}), ready, io.Discard, os.Stderr)
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
