package connect

import (
	"context"
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/alibaba/kt-connect/pkg/kt"
	"github.com/alibaba/kt-connect/pkg/kt/cluster"
	"github.com/alibaba/kt-connect/pkg/kt/dns"
	"github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/rs/zerolog/log"
	"strings"
)

func setupDns(cli kt.CliInterface, opt *options.DaemonOptions, shadowPodIp string) error {
	if strings.HasPrefix(opt.ConnectOptions.DnsMode, common.DnsModeHosts) {
		dump2HostsNamespaces := ""
		pos := len(common.DnsModeHosts)
		if len(opt.ConnectOptions.DnsMode) > pos + 1 && opt.ConnectOptions.DnsMode[pos:pos+1] == ":" {
			dump2HostsNamespaces = opt.ConnectOptions.DnsMode[pos+1:]
		}
		opt.RuntimeOptions.Dump2Host = setupDump2Host(cli.Kubernetes(), opt.Namespace,
			dump2HostsNamespaces, opt.ConnectOptions.ClusterDomain)
	} else if opt.ConnectOptions.DnsMode == common.DnsModePodDns {
		return dns.Ins().SetNameServer(cli.Kubernetes(), shadowPodIp, opt)
	} else if opt.ConnectOptions.DnsMode == common.DnsModeLocalDns {
		dnsPort := common.AlternativeDnsPort
		if util.IsWindows() {
			dnsPort = common.StandardDnsPort
		}
		if err := dns.SetupLocalDns(shadowPodIp, dnsPort); err != nil {
			log.Error().Err(err).Msgf("Failed to setup local dns server")
			return err
		}
		return dns.Ins().SetNameServer(cli.Kubernetes(), fmt.Sprintf("%s:%d", common.Localhost, dnsPort), opt)
	}
	return nil
}

func setupDump2Host(k cluster.KubernetesInterface, currentNamespace, targetNamespaces, clusterDomain string) bool {
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
		host := getServiceHosts(k, namespace)
		for svc, ip := range host {
			if ip == "" || ip == "None" {
				continue
			}
			log.Debug().Msgf("Service found: %s.%s %s", svc, namespace, ip)
			if namespace == currentNamespace {
				hosts[svc] = ip
			}
			hosts[svc+"."+namespace] = ip
			hosts[svc+"."+namespace+".svc."+clusterDomain] = ip
		}
	}
	return util.DumpHosts(hosts)
}

func getServiceHosts(k cluster.KubernetesInterface, namespace string) map[string]string {
	hosts := map[string]string{}
	services, err := k.GetAllServiceInNamespace(context.TODO(), namespace)
	if err == nil {
		for _, service := range services.Items {
			hosts[service.Name] = service.Spec.ClusterIP
		}
	}
	return hosts
}

func getOrCreateShadow(kubernetes cluster.KubernetesInterface, opt *options.DaemonOptions) (string, string, *util.SSHCredential, error) {
	shadowPodName := fmt.Sprintf("kt-connect-shadow-%s", strings.ToLower(util.RandomString(5)))
	if opt.ConnectOptions.SharedShadow {
		shadowPodName = fmt.Sprintf("kt-connect-shadow-daemon")
	}

	endPointIP, podName, credential, err := cluster.GetOrCreateShadow(context.TODO(), kubernetes,
		shadowPodName, opt, getLabels(shadowPodName), make(map[string]string), getEnvs(opt))
	if err != nil {
		return "", "", nil, err
	}

	return endPointIP, podName, credential, nil
}

func getEnvs(opt *options.DaemonOptions) map[string]string {
	envs := make(map[string]string)
	localDomains := dns.GetLocalDomains()
	if localDomains != "" {
		log.Debug().Msgf("Found local domains: %s", localDomains)
		envs[common.EnvVarLocalDomains] = localDomains
	}
	if opt.ConnectOptions.DnsMode == common.DnsModeLocalDns {
		envs[common.EnvVarDnsProtocol] = "tcp"
	} else {
		envs[common.EnvVarDnsProtocol] = "udp"
	}
	if opt.Debug {
		envs[common.EnvVarLogLevel] = "debug"
	} else {
		envs[common.EnvVarLogLevel] = "info"
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
