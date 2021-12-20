package connect

import (
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec/kubectl"
	"github.com/alibaba/kt-connect/pkg/kt/exec/portforward"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/registry"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/rs/zerolog/log"
)

func BySocks(cli kt.CliInterface, options *options.DaemonOptions) error {
	kubernetes, err := cli.Kubernetes()
	if err != nil {
		return err
	}
	options.RuntimeOptions.Dump2Host = setupDump2Host(kubernetes, options.Namespace,
		options.ConnectOptions.Dump2HostsNamespaces, options.ConnectOptions.ClusterDomain)

	if options.ConnectOptions.UseGlobalProxy {
		setupGlobalProxy(options)
	}

	_, podName, _, err := getOrCreateShadow(kubernetes, options)
	if err != nil {
		return err
	}

	return forwardSocksTunnelToLocal(cli.Exec().PortForward(), cli.Exec().Kubectl(), options, podName)
}

func setupGlobalProxy(options *options.DaemonOptions) {
	err := registry.SetGlobalProxy(options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to setup global connect proxy")
	}
	err = registry.SetHttpProxyEnvironmentVariable(options.ConnectOptions.Method, options.ConnectOptions.SocksPort, &options.RuntimeOptions.ProxyConfig)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to setup global http proxy")
	}
}

func forwardSocksTunnelToLocal(pfCli portforward.CliInterface, kubectlCli kubectl.CliInterface,
	options *options.DaemonOptions, podName string) (err error) {
	showSetupSocksMessage(common.ConnectMethodSocks, options.ConnectOptions)
	if options.UseKubectl {
		err = tunnel.PortForwardViaKubectl(kubectlCli, options, podName, common.Socks4Port, options.ConnectOptions.SocksPort)
	} else {
		_, _, err = pfCli.ForwardPodPortToLocal(options, podName, common.Socks4Port, options.ConnectOptions.SocksPort)
	}
	return err
}
