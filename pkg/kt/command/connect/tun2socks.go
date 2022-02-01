package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/tun"
	"github.com/alibaba/kt-connect/pkg/kt/tunnel"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func ByTun2Socks(cli kt.CliInterface, opt *options.DaemonOptions) error {
	podIP, podName, credential, err := getOrCreateShadow(cli.Kubernetes(), opt)
	if err != nil {
		return err
	}
	go activePodRoute(cli, opt, podName)

	_, _, err = tunnel.ForwardSSHTunnelToLocal(opt, podName, opt.ConnectOptions.SSHPort)
	if err != nil {
		return err
	}
	if err = startSocks5Connection(opt, credential.PrivateKeyPath); err != nil {
		return err
	}

	if opt.ConnectOptions.DisableTunDevice {
		showSetupSocksMessage(opt.ConnectOptions.SocksPort)
		if strings.HasPrefix(opt.ConnectOptions.DnsMode, common.DnsModeHosts) {
			return setupDns(cli, opt, podIP)
		} else {
			return nil
		}
	} else {
		if err = tun.Ins().CheckContext(); err != nil {
			return err
		}
		socksAddr := fmt.Sprintf("socks5://127.0.0.1:%d", opt.ConnectOptions.SocksPort)
		if err = tun.Ins().ToSocks(socksAddr, opt.Debug); err != nil {
			return err
		}
		log.Info().Msgf("Tun device %s is ready", tun.Ins().GetName())

		if !opt.ConnectOptions.DisableTunRoute {
			if err = setupTunRoute(cli, opt); err != nil {
				return err
			}
			log.Info().Msgf("Route to tun device completed")
		}
		return setupDns(cli, opt, podIP)
	}
}

func activePodRoute(cli kt.CliInterface, opt *options.DaemonOptions, podName string) {
	_, stderr, _ := cli.Kubernetes().ExecInPod(common.DefaultContainer, podName, opt.Namespace, *opt.RuntimeOptions,
		"nslookup", "kubernetes.default.svc")
	if stderr != "" {
		log.Debug().Msgf("Pod route not ready, %s", stderr)
	}
}

func setupTunRoute(cli kt.CliInterface, options *options.DaemonOptions) error {
	cidrs, err := cli.Kubernetes().ClusterCidrs(context.TODO(), options.Namespace, options.ConnectOptions)
	if err != nil {
		return err
	}

	err = tun.Ins().SetRoute(cidrs)
	if err != nil {
		if tun.IsAllRouteFailError(err) {
			if strings.Contains(err.(tun.AllRouteFailError).OriginalError().Error(), "exit status") {
				log.Warn().Msgf(err.Error())
			} else {
				log.Warn().Err(err.(tun.AllRouteFailError).OriginalError()).Msgf(err.Error())
			}
			return err
		}
		if strings.Contains(err.Error(), "exit status") {
			log.Warn().Msgf("Some route rule is not setup properly")
		} else {
			log.Warn().Err(err).Msgf("Some route rule is not setup properly")
		}
	}
	return nil
}

func startSocks5Connection(options *options.DaemonOptions, privateKey string) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		// will hang here if not error happen
		success <-sshchannel.Ins().StartSocks5Proxy(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SSHPort),
			fmt.Sprintf("127.0.0.1:%d", options.ConnectOptions.SocksPort),
		)
	}()
	return <-success
}

func showSetupSocksMessage(socksPort int) {
	log.Info().Msgf("--------------------------------------------------------------")
	if util.IsWindows() {
		if util.IsCmd() {
			log.Info().Msgf("Please setup proxy config by: set http_proxy=socks5://127.0.0.1:%d", socksPort)
		} else {
			log.Info().Msgf("Please setup proxy config by: $env:http_proxy=\"socks5://127.0.0.1:%d\"", socksPort)
		}
	} else {
		log.Info().Msgf("Please setup proxy config by: export http_proxy=socks5://127.0.0.1:%d", socksPort)
	}
	log.Info().Msgf("--------------------------------------------------------------")
}
