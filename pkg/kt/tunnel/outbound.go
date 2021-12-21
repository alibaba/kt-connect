package tunnel

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
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

// ForwardSSHTunnelToLocal mapping local port to shadow pod ssh port
func ForwardSSHTunnelToLocal(kubectlCli kubectl.CliInterface, options *options.DaemonOptions,
	podName string, localSSHPort int) (stop chan struct{}, rootCtx context.Context, err error) {
	if options.UseKubectl {
		err = PortForwardViaKubectl(kubectlCli, options, podName, common.SshPort, localSSHPort)
	} else {
		stop, rootCtx, err = ForwardPodPortToLocal(options, podName, common.SshPort, localSSHPort)
	}
	return stop, rootCtx, err
}

// PortForwardViaKubectl port forwarding via kubectl tool
func PortForwardViaKubectl(kubectlCli kubectl.CliInterface, options *options.DaemonOptions, podName string, remotePort, localPort int) error {
	command := kubectlCli.PortForward(options.Namespace, podName, remotePort, localPort)

	// If localSSHPort is in use by another process, return an error.
	ready := util.WaitPortBeReady(1, localPort)
	if ready {
		return fmt.Errorf("127.0.0.1:%d already in use", localPort)
	}

	err := util.BackgroundRun(command, fmt.Sprintf("forward %d to localhost:%d", remotePort, localPort))
	if err == nil {
		if !util.WaitPortBeReady(options.WaitTime, localPort) {
			err = fmt.Errorf("connect to port-forward failed")
		}
		util.SetupPortForwardHeartBeat(localPort)
	}
	return err
}

// ForwardPodPortToLocal port forwarding via k8s api
func ForwardPodPortToLocal(options *options.DaemonOptions, podName string, remotePort, localPort int) (chan struct{}, context.Context, error) {
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
		return nil, nil, fmt.Errorf("connect to port-forward failed")
	}
	util.SetupPortForwardHeartBeat(localPort)
	return stop, rootCtx, nil
}

// PortForward call port forward api
func portForward(options *options.DaemonOptions, podName string, remotePort, localPort int, stop chan struct{}) error {
	ready := make(chan struct{})
	apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", options.Namespace, podName)
	log.Debug().Msgf("Request port forward %d:%d to %s", localPort, remotePort, options.RuntimeOptions.RestConfig.Host)
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
