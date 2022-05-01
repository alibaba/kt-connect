package dns

import (
	"fmt"
	"github.com/alibaba/kt-connect/pkg/common"
	opt "github.com/alibaba/kt-connect/pkg/kt/command/options"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"time"
)

type DnsServer struct {
	clusterDnsAddr string
	upstreamDnsAddr string
}

func SetupLocalDns(remoteDnsPort, localDnsPort int) error {
	var success = make(chan error)
	go func() {
		time.Sleep(1 * time.Second)
		success <-nil
	}()
	go func() {
		nameserver := GetNameServer()
		log.Debug().Msgf("Setup local DNS with shadow pod %s:%d and upstream %s:%d",
			common.Localhost, remoteDnsPort, nameserver, common.StandardDnsPort)
		success <-common.SetupDnsServer(&DnsServer{
			clusterDnsAddr: fmt.Sprintf("%s:%d", common.Localhost, remoteDnsPort),
			upstreamDnsAddr: fmt.Sprintf("%s:%d", nameserver, common.StandardDnsPort),
		}, localDnsPort, "udp")
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

	answer := common.ReadCache(domain, qtype, opt.Get().ConnectOptions.DnsCacheTtl)
	if answer != nil {
		log.Debug().Msgf("Found domain %s (%d) in cache", domain, qtype)
		return answer
	}

	res, err := common.NsLookup(domain, qtype, "tcp", clusterDnsAddr)
	if res != nil && len(res.Answer) > 0 {
		// only record none-empty result of cluster dns
		log.Debug().Msgf("Found domain %s (%d) in cluster dns (%s)", domain, qtype, clusterDnsAddr)
		common.WriteCache(domain, qtype, res.Answer)
		return res.Answer
	} else {
		if err != nil && !common.IsDomainNotExist(err) {
			// usually io timeout error
			log.Debug().Err(err).Msgf("Failed to lookup %s (%d) in cluster dns (%s)", domain, qtype, clusterDnsAddr)
		}

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
	if err == nil || common.IsDomainNotExist(err) {
		common.WriteCache(domain, qtype, []dns.RR{})
	}
	return []dns.RR{}
}
