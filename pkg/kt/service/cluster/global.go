package cluster

import (
	"context"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/options"
	"github.com/rs/zerolog/log"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ResourceMeta ...
type ResourceMeta struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

// SSHkeyMeta ...
type SSHkeyMeta struct {
	SshConfigMapName string
	PrivateKeyPath   string
}

// ClusterCidrs get cluster Cidrs
func (k *Kubernetes) ClusterCidrs(namespace string) (cidrs []string, err error) {
	if !opt.Get().ConnectOptions.DisablePodIp {
		cidrs, err = getPodCidrs(k.Clientset, namespace)
		if err != nil {
			return
		}
	}
	log.Debug().Msgf("Pod CIDR is %v", cidrs)

	serviceCidr, err := getServiceCidr(k.Clientset, namespace)
	if err != nil {
		return
	}
	cidrs = append(cidrs, serviceCidr...)
	log.Debug().Msgf("Service CIDR is %v", serviceCidr)

	if opt.Get().ConnectOptions.IncludeIps != "" {
		for _, ipRange := range strings.Split(opt.Get().ConnectOptions.IncludeIps, ",") {
			if opt.Get().ConnectOptions.Mode == common.ConnectModeTun2Socks && isSingleIp(ipRange) {
				log.Warn().Msgf("Includes single IP '%s' is not allow in %s mode", ipRange, common.ConnectModeTun2Socks)
			} else {
				cidrs = append(cidrs, ipRange)
			}
		}
	}
	return
}

// GetAllNamespaces get all namespaces
func (k *Kubernetes) GetAllNamespaces() (*coreV1.NamespaceList, error) {
	return k.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
}
