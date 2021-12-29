package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func setupDump2Host(kubernetes cluster.KubernetesInterface, currentNamespace, targetNamespaces, clusterDomain string) bool {
	namespacesToDump := []string{currentNamespace}
	if targetNamespaces != "" {
		namespacesToDump = []string{}
		for _, ns := range strings.Split(targetNamespaces, ",") {
			namespacesToDump = append(namespacesToDump, ns)
		}
	}
	hosts := map[string]string{}
	for _, namespace := range namespacesToDump {
		log.Debug().Msgf("Search service in %s namespace ...", namespace)
		host := kubernetes.GetServiceHosts(context.TODO(), namespace)
		for svc, ip := range host {
			if ip == "" || ip == "None" {
				continue
			}
			log.Debug().Msgf("Service found: %s.%s %s", svc, namespace, ip)
			if namespace == currentNamespace {
				hosts[svc] = ip
			}
			hosts[svc+"."+namespace] = ip
			hosts[svc+"."+namespace+"."+clusterDomain] = ip
		}
	}
	return util.DumpHosts(hosts)
}

func getOrCreateShadow(kubernetes cluster.KubernetesInterface, options *options.DaemonOptions) (string, string, *util.SSHCredential, error) {
	shadowPodName := fmt.Sprintf("kt-connect-shadow-%s", strings.ToLower(util.RandomString(5)))
	if options.ConnectOptions.ShareShadow {
		shadowPodName = fmt.Sprintf("kt-connect-shadow-daemon")
	}

	endPointIP, podName, credential, err := cluster.GetOrCreateShadow(context.TODO(), kubernetes,
		shadowPodName, options, getLabels(shadowPodName), make(map[string]string), getEnvs(options))
	if err != nil {
		return "", "", nil, err
	}

	return endPointIP, podName, credential, nil
}

func showSetupSocksMessage(protocol string, connectOptions *options.ConnectOptions) {
	port := connectOptions.SocksPort
	log.Info().Msgf("Starting up %s proxy ...", protocol)
	if !connectOptions.UseGlobalProxy {
		log.Info().Msgf("--------------------------------------------------------------")
		if util.IsWindows() {
			if util.IsCmd() {
				log.Info().Msgf("Please setup proxy config by: set http_proxy=%s://127.0.0.1:%d", protocol, port)
			} else {
				log.Info().Msgf("Please setup proxy config by: $env:http_proxy=\"%s://127.0.0.1:%d\"", protocol, port)
			}
		} else {
			log.Info().Msgf("Please setup proxy config by: export http_proxy=%s://127.0.0.1:%d", protocol, port)
		}
		log.Info().Msgf("--------------------------------------------------------------")
	}
}

func getEnvs(options *options.DaemonOptions) map[string]string {
	envs := make(map[string]string)
	localDomains := util.GetLocalDomains()
	if localDomains != "" {
		log.Debug().Msgf("Found local domains: %s", localDomains)
		envs[common.EnvVarLocalDomains] = localDomains
	}
	if options.ConnectOptions.Method == common.ConnectMethodTun {
		envs[common.ClientTunIP] = options.RuntimeOptions.SourceIP
		envs[common.ServerTunIP] = options.RuntimeOptions.DestIP
		envs[common.TunMaskLength] = util.ExtractNetMaskFromCidr(options.ConnectOptions.TunCidr)
	}
	return envs
}

func getLabels(workload string) map[string]string {
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RoleConnectShadow,
		common.KtName:    workload,
	}
	return labels
}
