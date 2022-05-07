package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/alibaba/kt-connect/pkg/kt/util"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type DnsServer struct {
	dnsAddresses []string
}

func SetupLocalDns(remoteDnsPort, localDnsPort int, dnsOrder []string) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		upstreamDnsAddresses := getDnsAddresses(dnsOrder, remoteDnsPort)
		log.Info().Msgf("Setup local DNS with upstream %v", upstreamDnsAddresses)
		success <- common.SetupDnsServer(&DnsServer{upstreamDnsAddresses}, localDnsPort, "udp")
	}()
	return <-success
}

func getDnsAddresses(dnsOrder []string, clusterDnsPort int) []string {
	var dnsAddresses []string
	for _, dnsAddr := range dnsOrder {
		switch dnsAddr {
		case util.DnsOrderCluster:
			dnsAddresses = append(dnsAddresses, fmt.Sprintf("tcp:%s:%d", common.Localhost, clusterDnsPort))
		case util.DnsOrderUpstream:
			upstreamDns := GetNameServer()
			if upstreamDns != "" {
				dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", upstreamDns, common.StandardDnsPort))
			}
		default:
			switch strings.Count(dnsAddr, ":") {
			case 0:
				dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s:%d", dnsAddr, common.StandardDnsPort))
			case 1:
				dnsAddresses = append(dnsAddresses, fmt.Sprintf("udp:%s", dnsAddr))
			case 2:
				dnsAddresses = append(dnsAddresses, dnsAddr)
			default:
				log.Warn().Msgf("Skip invalid dns server %s", dnsAddr)
			}
		}
	}
	return dnsAddresses
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	msg.Answer = query(req, s.dnsAddresses)
	if err := w.WriteMsg(msg); err != nil {
		log.Warn().Err(err).Msgf("Failed to reply dns request")
	}
}

func query(req *dns.Msg, dnsAddresses []string) []dns.RR {
	domain := req.Question[0].Name
	qtype := req.Question[0].Qtype

	answer := common.ReadCache(domain, qtype, opt.Get().ConnectOptions.DnsCacheTtl)
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", domain, qtype)
		return answer
	}

	for _, dnsAddr := range dnsAddresses {
		dnsParts := strings.SplitN(dnsAddr, ":", 2)
		protocol := dnsParts[0]
		ipAndPort := dnsParts[1]
		if strings.HasPrefix(ipAndPort, ":") || strings.Count(ipAndPort, ":") != 1 {
			// skip item with empty or invalid ip address
			continue
		}
		res, err := common.NsLookup(domain, qtype, protocol, ipAndPort)
		if res != nil && len(res.Answer) > 0 {
			// only record none-empty result of cluster dns
			log.Debug().Msgf("Found domain %s (%d) in dns (%s)", domain, qtype, ipAndPort)
			common.WriteCache(domain, qtype, res.Answer)
			return res.Answer
		} else if err != nil && !common.IsDomainNotExist(err) {
			// usually io timeout error
			log.Warn().Err(err).Msgf("Failed to lookup %s (%d) in dns (%s)", domain, qtype, ipAndPort)
		}
	}
	log.Debug().Msgf("Empty answer for domain lookup %s (%d)", domain, qtype)
	common.WriteCache(domain, qtype, []dns.RR{})
	return []dns.RR{}
}
