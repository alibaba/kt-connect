package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"time"
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
	if err = startSocks5Connection(cli.Exec().SshChannel(), options, credential.PrivateKeyPath); err != nil {
		return err
	}

	socksAddr := fmt.Sprintf("socks5://%s:%d", options.ConnectOptions.SocksAddr, options.ConnectOptions.SocksPort)
	err = cli.Exec().Tunnel().ToSocks(socksAddr)
	if err != nil {
		return err
	}

	cidrs, err := cli.Kubernetes().ClusterCidrs(context.TODO(), options.Namespace, options.ConnectOptions)
	if err != nil {
		return err
	}

	if err = cli.Exec().Tunnel().SetRoute(cidrs); err != nil {
		return err
	}

	return nil
}

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions, privateKey string) (err error) {
	jvmrcFilePath := util.GetJvmrcFilePath(options.ConnectOptions.JvmrcDir)
	if jvmrcFilePath != "" {
		ioutil.WriteFile(jvmrcFilePath, []byte(fmt.Sprintf("-DsocksProxyHost=127.0.0.1\n-DsocksProxyPort=%d",
			options.ConnectOptions.SocksPort)), 0644)
	}

	var success = make(chan bool)
	go func() {
		time.Sleep(2 * time.Second)
		success <- true
	}()
	go func() {
		err = ssh.StartSocks5Proxy(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
			fmt.Sprintf("%s:%d", options.ConnectOptions.SocksAddr, options.ConnectOptions.SocksPort),
		)
		if err != nil {
			log.Error().Err(err).Msgf("failed to create socks5 connection")
			success <- false
		}
	}()
	if <- success {
		showSetupSocksMessage(common.ConnectMethodSocks5, options.ConnectOptions)
	}
	return
}
