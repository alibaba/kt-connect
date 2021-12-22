package connect

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"io/ioutil"
)

func BySocks5(cli kt.CliInterface, options *options.DaemonOptions) error {
	options.RuntimeOptions.Dump2Host = setupDump2Host(cli.Kubernetes(), options.Namespace,
		options.ConnectOptions.Dump2HostsNamespaces, options.ConnectOptions.ClusterDomain)

	_, podName, credential, err := getOrCreateShadow(cli.Kubernetes(), options)
	if err != nil {
		return err
	}

	_, _, err = tunnel.ForwardSSHTunnelToLocal(cli.Exec().Kubectl(), options, podName, options.ConnectOptions.SSHPort)
	if err != nil {
		return err
	}
	return startSocks5Connection(cli.Exec().SshChannel(), options, credential.PrivateKeyPath)
}

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions, privateKey string) (err error) {
	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		ioutil.WriteFile(jvmrcFilePath, []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
			options.ConnectOptions.SocksPort)), 0644)
	}

	showSetupSocksMessage(common.ConnectMethodSocks5, options.ConnectOptions)
	return ssh.StartSocks5Proxy(
		privateKey,
		fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
		fmt.Sprintf("%s:%d", options.ConnectOptions.SocksAddr, options.ConnectOptions.SocksPort),
	)
}
