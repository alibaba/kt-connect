package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	cluster2 "github.com/alibaba/kt-connect/pkg/kt/service/cluster"
	dns2 "github.com/alibaba/kt-connect/pkg/kt/service/dns"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func setupDns(shadowPodIp string) error {
	if strings.HasPrefix(opt.Get().ConnectOptions.DnsMode, common.DnsModeHosts) {
		dump2HostsNamespaces := ""
		pos := len(common.DnsModeHosts)
		if len(opt.Get().ConnectOptions.DnsMode) > pos + 1 && opt.Get().ConnectOptions.DnsMode[pos:pos+1] == ":" {
			dump2HostsNamespaces = opt.Get().ConnectOptions.DnsMode[pos+1:]
		}
		if err := dumpToHost(opt.Get().Namespace, dump2HostsNamespaces, opt.Get().ConnectOptions.ClusterDomain); err != nil {
			return err
		}
	} else if opt.Get().ConnectOptions.DnsMode == common.DnsModePodDns {
		return dns2.Ins().SetNameServer(shadowPodIp)
	} else if opt.Get().ConnectOptions.DnsMode == common.DnsModeLocalDns {
		if err := dumpCurrentNamespaceToHost(opt.Get().Namespace); err != nil {
			return err
		}
		dnsPort := common.AlternativeDnsPort
		if util.IsWindows() {
			dnsPort = common.StandardDnsPort
		}
		// must setup name server before change dns config
		// otherwise the upstream name server address will be incorrect in linux
		if err := dns2.SetupLocalDns(shadowPodIp, dnsPort); err != nil {
			log.Error().Err(err).Msgf("Failed to setup local dns server")
			return err
		}
		return dns2.Ins().SetNameServer(fmt.Sprintf("%s:%d", common.Localhost, dnsPort))
	} else {
		return fmt.Errorf("invalid dns mode: '%s', supportted mode are %s, %s, %s", opt.Get().ConnectOptions.DnsMode,
			common.DnsModeLocalDns, common.DnsModePodDns, common.DnsModeHosts)
	}
	return nil
}

func dumpToHost(currentNamespace, targetNamespaces, clusterDomain string) error {
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
		svcToIp := getServiceHosts(namespace)
		for svc, ip := range svcToIp {
			if namespace == currentNamespace {
				hosts[svc] = ip
			}
			hosts[svc+"."+namespace] = ip
			hosts[svc+"."+namespace+".svc."+clusterDomain] = ip
		}
	}
	return dns2.DumpHosts(hosts)
}

func dumpCurrentNamespaceToHost(currentNamespace string) error {
	log.Debug().Msgf("Search service in %s namespace ...", currentNamespace)
	return dns2.DumpHosts(getServiceHosts(currentNamespace))
}

func getServiceHosts(namespace string) map[string]string {
	hosts := map[string]string{}
	services, err := cluster2.Ins().GetAllServiceInNamespace(context.TODO(), namespace)
	if err == nil {
		for _, service := range services.Items {
			ip := service.Spec.ClusterIP
			if ip == "" || ip == "None" {
				pods, err2 := cluster2.Ins().GetPodsByLabel(context.TODO(), service.Spec.Selector, namespace)
				if err2 != nil || len(pods.Items) == 0 {
					continue
				}
				ip = pods.Items[0].Status.PodIP
				log.Debug().Msgf("Headless service found: %s.%s %s", service.Name, namespace, ip)
			} else {
				log.Debug().Msgf("Service found: %s.%s %s", service.Name, namespace, ip)
			}
			hosts[service.Name] = ip
		}
	}
	return hosts
}

func getOrCreateShadow() (string, string, string, error) {
	shadowPodName := fmt.Sprintf("kt-connect-shadow-%s", strings.ToLower(util.RandomString(5)))
	if opt.Get().ConnectOptions.SharedShadow {
		shadowPodName = fmt.Sprintf("kt-connect-shadow-daemon")
	}

	endPointIP, podName, privateKeyPath, err := cluster2.GetOrCreateShadow(context.TODO(),
		shadowPodName, getLabels(), make(map[string]string), getEnvs())
	if err != nil {
		return "", "", "", err
	}

	return endPointIP, podName, privateKeyPath, nil
}

func getEnvs() map[string]string {
	envs := make(map[string]string)
	localDomains := dns2.GetLocalDomains()
	if localDomains != "" {
		log.Debug().Msgf("Found local domains: %s", localDomains)
		envs[common.EnvVarLocalDomains] = localDomains
	}
	if opt.Get().ConnectOptions.DnsMode == common.DnsModeLocalDns {
		envs[common.EnvVarDnsProtocol] = "tcp"
	} else {
		envs[common.EnvVarDnsProtocol] = "udp"
	}
	if opt.Get().Debug {
		envs[common.EnvVarLogLevel] = "debug"
	} else {
		envs[common.EnvVarLogLevel] = "info"
	}
	return envs
}

func getLabels() map[string]string {
	labels := map[string]string{
		common.ControlBy: common.KubernetesTool,
		common.KtRole:    common.RoleConnectShadow,
	}
	return labels
}
