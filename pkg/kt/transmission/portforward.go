package transmission

import (
	"fmt"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// SetupPortForwardToLocal mapping local port to shadow pod ssh port
func SetupPortForwardToLocal(podName string, remotePort, localPort int) error {
	return setupPortForwardToLocal(podName, remotePort, localPort, true)
}

func setupPortForwardToLocal(podName string, remotePort, localPort int, isInitConnect bool) error {
	ready := make(chan struct{})
	var ticker *time.Ticker
	go func() {
		if err := portForward(podName, remotePort, localPort, ready); err != nil {
			if isInitConnect {
				log.Error().Err(err).Msgf("Failed to setup port forward local:%d -> pod %s:%d", localPort, podName, remotePort)
			} else {
				log.Debug().Err(err).Msgf("Port forward local:%d -> pod %s:%d interrupted", localPort, podName, remotePort)
			}
			time.Sleep(time.Duration(opt.Get().Global.PortForwardTimeout) * time.Second)
		}
		if ticker != nil {
			ticker.Stop()
		}
		log.Debug().Msgf("Port forward reconnecting ...")
		_ = setupPortForwardToLocal(podName, remotePort, localPort, false)
	}()

	select {
	case <-ready:
		ticker = cluster.SetupPortForwardHeartBeat(localPort)
		log.Info().Msgf("Port forward local:%d -> pod %s:%d established", localPort, podName, remotePort)
		return nil
	case <-time.After(time.Duration(opt.Get().Global.PortForwardTimeout) * time.Second):
		return fmt.Errorf("connect to port-forward failed")
	}
}

// PortForward call port forward api
func portForward(podName string, remotePort, localPort int, ready chan struct{}) error {
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", opt.Get().Global.Namespace, podName)
	log.Debug().Msgf("Request port forward pod:%d -> local:%d via %s", remotePort, localPort, opt.Store.RestConfig.Host)
	apiUrl, err := parseReqHost(opt.Store.RestConfig.Host, apiPath)

	transport, upgrader, err := spdy.RoundTripperFor(opt.Store.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiUrl)
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	fw, err := portforward.New(dialer, ports, make(<-chan struct{}), ready, util.BackgroundLogger, util.BackgroundLogger)
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
