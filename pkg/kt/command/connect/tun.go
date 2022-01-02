package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/exec/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"time"
)

func ByTun(cli kt.CliInterface, options *options.DaemonOptions) error {
	options.RuntimeOptions.Dump2Host = setupDump2Host(cli.Kubernetes(), options.Namespace,
		options.ConnectOptions.Dump2HostsNamespaces, options.ConnectOptions.ClusterDomain)

	_, podName, credential, err := getOrCreateShadow(cli.Kubernetes(), options)
	if err != nil {
		return err
	}

	_, _, err = tunnel.ForwardSSHTunnelToLocal(options, podName, options.ConnectOptions.SSHPort)
	if err != nil {
		return err
	}
	if err = startSocks5Connection(cli.Exec().SshChannel(), options, credential.PrivateKeyPath); err != nil {
		return err
	}

	socksAddr := fmt.Sprintf("socks5://127.0.0.1:%d", options.ConnectOptions.SocksPort)
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

func startSocks5Connection(ssh sshchannel.Channel, options *options.DaemonOptions, privateKey string) error {
	var success = make(chan error)
	go func() {
		time.Sleep(2 * time.Second)
		success <-nil
	}()
	go func() {
		// will hang here if not error happen
		success <-ssh.StartSocks5Proxy(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
			fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
		)
	}()
	return <-success
}
