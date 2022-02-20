package connect

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/service/sshchannel"
	"github.com/alibaba/kt-connect/pkg/kt/service/tun"
	"github.com/alibaba/kt-connect/pkg/kt/transmission"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func ByTun2Socks() error {
	podIP, podName, privateKeyPath, err := getOrCreateShadow()
	if err != nil {
		return err
	}
	go activePodRoute(podName)

	localSshPort, err := util.GetRandomTcpPort()
	if err != nil {
		return err
	}
	if _, err = transmission.SetupPortForwardToLocal(podName, common.StandardSshPort, localSshPort); err != nil {
		return err
	}
	if err = startSocks5Connection(privateKeyPath, localSshPort); err != nil {
		return err
	}

	if opt.Get().ConnectOptions.DisableTunDevice {
		showSetupSocksMessage(opt.Get().ConnectOptions.SocksPort)
		if strings.HasPrefix(opt.Get().ConnectOptions.DnsMode, util.DnsModeHosts) {
			return setupDns(podName, podIP)
		} else {
			return nil
		}
	} else {
		if err = tun.Ins().CheckContext(); err != nil {
			return err
		}
		socksAddr := fmt.Sprintf("socks5://127.0.0.1:%d", opt.Get().ConnectOptions.SocksPort)
		if err = tun.Ins().ToSocks(socksAddr); err != nil {
			return err
		}
		log.Info().Msgf("Tun device %s is ready", tun.Ins().GetName())

		if !opt.Get().ConnectOptions.DisableTunRoute {
			if err = setupTunRoute(); err != nil {
				return err
			}
			log.Info().Msgf("Route to tun device completed")
		}
		return setupDns(podName, podIP)
	}
}

func activePodRoute(podName string) {
	stdout, stderr, err := cluster.Ins().ExecInPod(util.DefaultContainer, podName, opt.Get().Namespace,
		"nslookup", "-vc", "kubernetes.default.svc")
	log.Debug().Msgf("Active DNS %s", strings.TrimSpace(strings.Split(stdout, "\n")[0]))
	if stderr != "" || err != nil {
		log.Warn().Msgf("Pod route not ready yet, %s", stderr)
	}
}

func setupTunRoute() error {
	cidrs, err := cluster.Ins().ClusterCidrs(opt.Get().Namespace)
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

func startSocks5Connection(privateKey string, localSshPort int) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		// will hang here if not error happen
		success <- sshchannel.Ins().StartSocks5Proxy(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", localSshPort),
			fmt.Sprintf("127.0.0.1:%d", opt.Get().ConnectOptions.SocksPort),
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
