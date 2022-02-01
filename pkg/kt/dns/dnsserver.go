package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"time"
)

type DnsServer struct {
	clusterDnsAddr string
	upstreamDnsAddr string
}

func SetupLocalDns(upstreamIp string, dnsPort int) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		success <-common.SetupDnsServer(&DnsServer{
			clusterDnsAddr: fmt.Sprintf("%s:%d", upstreamIp, common.StandardDnsPort),
			upstreamDnsAddr: fmt.Sprintf("%s:%d", GetNameServer(), common.StandardDnsPort),
		}, dnsPort, "udp")
	}()
	return <-success
}

// ServeDNS query DNS record
func (s *DnsServer) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	msg := (&dns.Msg{}).SetReply(req)
	msg.Authoritative = true
	msg.Answer = query(req, s.clusterDnsAddr, s.upstreamDnsAddr)
	if err := w.WriteMsg(msg); err != nil {
		log.Warn().Err(err).Msgf("Failed to reply dns request")
	}
}

func query(req *dns.Msg, clusterDnsAddr, upstreamDnsAddr string) []dns.RR {
	domain := req.Question[0].Name
	qtype := req.Question[0].Qtype

	answer := common.ReadCache(domain, qtype)
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", domain, qtype)
		return answer
	}

	res, err := common.NsLookup(domain, qtype, "tcp", clusterDnsAddr)
	if err != nil && !common.IsDomainNotExist(err) {
		log.Warn().Err(err).Msgf("Failed to lookup %s (%d) in cluster dns (%s)", domain, qtype, clusterDnsAddr)
	} else if res != nil && len(res.Answer) > 0 {
		log.Debug().Msgf("Found domain %s (%d) in cluster dns (%s)", domain, qtype, clusterDnsAddr)
		common.WriteCache(domain, qtype, res.Answer)
		return res.Answer
	} else {
		res, err = common.NsLookup(domain, qtype, "udp", upstreamDnsAddr)
		if err != nil {
			if common.IsDomainNotExist(err) {
				log.Debug().Msgf(err.Error())
			} else {
				log.Warn().Err(err).Msgf("Failed to lookup %s (%d) in upstream dns (%s)", domain, qtype, upstreamDnsAddr)
			}
		} else if len(res.Answer) > 0 {
			log.Debug().Msgf("Found domain %s (%d) in upstream dns (%s)", domain, qtype, upstreamDnsAddr)
			common.WriteCache(domain, qtype, res.Answer)
			return res.Answer
		} else {
			log.Debug().Msgf("Empty answer for domain lookup %s (%d)", domain, qtype)
		}
	}
	common.WriteCache(domain, qtype, []dns.RR{})
	return []dns.RR{}
}
