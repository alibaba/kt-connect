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
		if err := dumpToHost(cli.Kubernetes(), opt.Namespace, dump2HostsNamespaces, opt.ConnectOptions.ClusterDomain); err != nil {
			return err
		}
	} else if opt.ConnectOptions.DnsMode == common.DnsModePodDns {
		return dns.Ins().SetNameServer(cli.Kubernetes(), shadowPodIp, opt)
	} else if opt.ConnectOptions.DnsMode == common.DnsModeLocalDns {
		if err := dumpCurrentNamespaceToHost(cli.Kubernetes(), opt.Namespace); err != nil {
			return err
		}
		dnsPort := common.AlternativeDnsPort
		if util.IsWindows() {
			dnsPort = common.StandardDnsPort
		}
		// must setup name server before change dns config
		// otherwise the upstream name server address will be incorrect in linux
		if err := dns.SetupLocalDns(shadowPodIp, dnsPort); err != nil {
			log.Error().Err(err).Msgf("Failed to setup local dns server")
			return err
		}
		return dns.Ins().SetNameServer(cli.Kubernetes(), fmt.Sprintf("%s:%d", common.Localhost, dnsPort), opt)
	} else {
		return fmt.Errorf("invalid dns mode: '%s', supportted mode are %s, %s, %s", opt.ConnectOptions.DnsMode,
			common.DnsModeLocalDns, common.DnsModePodDns, common.DnsModeHosts)
	}
	return nil
}

func dumpToHost(k cluster.KubernetesInterface, currentNamespace, targetNamespaces, clusterDomain string) error {
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
	return dns.DumpHosts(hosts)
}

func dumpCurrentNamespaceToHost(k cluster.KubernetesInterface, currentNamespace string) error {
	hosts := map[string]string{}
	log.Debug().Msgf("Search service in %s namespace ...", currentNamespace)
	host := getServiceHosts(k, currentNamespace)
	for svc, ip := range host {
		if ip == "" || ip == "None" {
			continue
		}
		log.Debug().Msgf("Service found: %s.%s %s", svc, currentNamespace, ip)
		hosts[svc] = ip
	}
	return dns.DumpHosts(hosts)
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
