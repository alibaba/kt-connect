package connect

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
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

	localSshPort := util.GetRandomTcpPort()
	if err = transmission.SetupPortForwardToLocal(podName, common.StandardSshPort, localSshPort); err != nil {
		return err
	}
	if err = startSocks5Connection(privateKeyPath, localSshPort); err != nil {
		return err
	}

	if opt.Get().ConnectOptions.DisableTunDevice {
		showSetupSocksMessage(opt.Get().ConnectOptions.SocksPort)
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
	}
	return setupDns(podName, podIP)
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
	var res = make(chan error)
	gone := false
	go func() {
		// will hang here if not error happen
		err := sshchannel.Ins().StartSocks5Proxy(
			privateKey,
			fmt.Sprintf("127.0.0.1:%d", localSshPort),
			fmt.Sprintf("127.0.0.1:%d", opt.Get().ConnectOptions.SocksPort),
		)
		log.Warn().Err(err).Msgf("Socks proxy broken")
		if !gone {
			res <-err
		}
		time.Sleep(10 * time.Second)
		log.Debug().Msgf("Socks proxy reconnecting ...")
		_ = startSocks5Connection(privateKey, localSshPort)
	}()
	select {
	case err := <-res:
		return err
	case <-time.After(1 * time.Second):
		gone = true
		return nil
	}
}

func showSetupSocksMessage(socksPort int) {
	if util.IsWindows() {
		if util.IsCmd() {
			log.Info().Msgf(">> Please setup proxy config by: set http_proxy=socks5://127.0.0.1:%d <<", socksPort)
		} else {
			log.Info().Msgf(">> Please setup proxy config by: $env:http_proxy=\"socks5://127.0.0.1:%d\" <<", socksPort)
		}
	} else {
		log.Info().Msgf(">> Please setup proxy config by: export http_proxy=socks5://127.0.0.1:%d <<", socksPort)
	}
}
